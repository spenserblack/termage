package gif

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"time"
)

var (
	// ErrAnimationComplete signifies that the animation should not continue.
	ErrAnimationComplete = errors.New("Animation is complete")
	alpha                = color.RGBA{0, 0, 0, 0}
)

// Frame is a helper struct to group together GIF frame info, which is stored
// in 3 separate slices by the standard library.
type Frame struct {
	image.Image
	delay          int
	disposalMethod byte
}

// Helper simplifies interacting with an animated GIF.
type Helper struct {
	// Frames is the group of frames composing the GIF.
	Frames    []Frame
	loopCount looper
	index     int
}

// NewHelper constructs a helper for managing animated GIFs.
func NewHelper(g *gif.GIF) (helper Helper, err error) {
	if !IsAnimated(g) {
		return helper, errors.New("GIF isn't animated")
	}
	if len(g.Delay) < len(g.Image) {
		return helper, errors.New("Not enough delays")
	}
	if len(g.Disposal) < len(g.Image) {
		return helper, errors.New("Not enough disposals")
	}
	frames := make([]Frame, 1, len(g.Image))
	frames[0] = Frame{g.Image[0], g.Delay[0], g.Disposal[0]}
	for i, v := range g.Image[1:] {
		prevFrame := frames[len(frames)-1]
		nextFrame := image.NewRGBA(prevFrame.Bounds())
		draw.Src.Draw(
			nextFrame,
			nextFrame.Bounds(),
			prevFrame,
			image.Point{},
		)
		draw.Over.Draw(nextFrame, nextFrame.Bounds(), v, image.Point{})
		frames = append(frames, Frame{
			nextFrame,
			g.Delay[i+1],
			g.Disposal[i+1],
		})
	}

	var l looper
	switch g.LoopCount {
	case -1:
		l = noLoop{}
	case 0:
		l = infiniteLoop{}
	default:
		l = &countLoop{g.LoopCount}
	}

	helper = Helper{
		frames,
		l,
		0,
	}
	return
}

// HelperFromReader cretes a new GIF helper from a Reader.
func HelperFromReader(r io.Reader) (helper Helper, err error) {
	var g *gif.GIF
	g, err = gif.DecodeAll(r)
	if err != nil {
		return
	}
	helper, err = NewHelper(g)
	return
}

// Delay returns the delay of the current frame.
func (h Helper) Delay() time.Duration {
	return time.Duration(h.Frames[h.index].delay) * (time.Second / 100)
}

// NextFrame moves along to the next frame and generates a new current image.
//
// It can return ErrAnimationComplete if the animation is complete and a new image
// does not need to be generated.
func (h *Helper) NextFrame() error {
	h.index++
	if h.index >= len(h.Frames) {
		if err := h.loopCount.nextLoop(); err != nil {
			return err
		}
		h.index = 0
		return nil
	}
	return nil
}

// CurrentImage is the image representing the current state of the animation.
func (h *Helper) CurrentImage() image.Image {
	return h.Frames[h.index]
}

// Looper is a helper to determine if looping should continue.
type looper interface {
	// NextLoop should be called at the end of each loop,
	// to determine if it should continue.
	nextLoop() error
}

// InfiniteLoop will never end animation.
type infiniteLoop struct{}

// NoLoop will always end animation.
type noLoop struct{}

// countLoop will end animation once the set number of loops completes.
type countLoop struct {
	count int
}

func (l infiniteLoop) nextLoop() error {
	return nil
}

func (l noLoop) nextLoop() error {
	return ErrAnimationComplete
}

func (l *countLoop) nextLoop() error {
	l.count--
	if l.count < 0 {
		return ErrAnimationComplete
	}
	return nil
}

// IsAnimated checks if the GIF should be animated.
func IsAnimated(g *gif.GIF) bool {
	return len(g.Image) > 1
}
