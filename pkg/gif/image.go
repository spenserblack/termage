package gif

import (
	"image"
	"image/color"
)

// ColorModel returns the color model of the image representing the
// current state of the animation.
func (h *Helper) ColorModel() color.Model {
	return h.CurrentImage().ColorModel()
}

// Bounds returns the bounds of the background/first frame of the GIF.
func (h *Helper) Bounds() image.Rectangle {
	return h.frames[0].Bounds()
}

// At returns the color of the pixel at (x, y) at the current state of
// the animation. In other words, the pixel that results from layering all
// GIF frames up to the current frame on top of each other.
func (h *Helper) At(x, y int) color.Color {
	return h.CurrentImage().At(x, y)
}
