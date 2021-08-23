package gif

import (
	"errors"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestInvalidReader checks that an error will be returned if the reader cannot
// be decoded to a GIF.
func TestInvalidReader(t *testing.T) {
	_, err := HelperFromReader(alwaysErrReader{})
	if err == nil {
		t.Errorf(`err = nil`)
	}
}

// TestAnimationIsAnimated checks that IsAnimated returns true for an animated GIF.
func TestAnimationIsAnimated(t *testing.T) {
	f, err := os.Open(getResource("spinning-2x2.gif"))
	g, err := gif.DecodeAll(f)
	if err != nil {
		panic(err)
	}
	if !IsAnimated(g) {
		t.Fatal(`IsAnimated = false, want true`)
	}
}

// TestSpinning2x2 loads an animated GIF and checks that it loops forever and
// the frames have the expected colors.
func TestSpinning2x2(t *testing.T) {
	f, err := os.Open(getResource("spinning-2x2.gif"))
	if err != nil {
		panic(err)
	}
	gifHelper, err := HelperFromReader(f)
	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}

	if _, ok := gifHelper.loopCount.(infiniteLoop); !ok {
		t.Errorf(`loopCount is %T, want infiniteLoop`, gifHelper.loopCount)
	}
	if l := len(gifHelper.Frames); l != 4 {
		t.Fatalf(`%d frames, want 4`, l)
	}
	bounds := gifHelper.Bounds()
	if want := (image.Point{0, 0}); bounds.Min != want {
		t.Fatalf(`Bounds().Min = %v, want %v`, bounds.Min, want)
	}
	if want := (image.Point{2, 2}); bounds.Max != want {
		t.Fatalf(`Bounds().Max = %v, want %v`, bounds.Max, want)
	}

	wantDelays := []time.Duration{
		500 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
	}

	backgroundWantColors := [][]color.Color{
		{color.White, color.White},
		{color.Black, color.White},
	}
	frame1WantColors := [][]color.Color{
		{color.Black, color.White},
		{color.White, color.White},
	}
	frame2WantColors := [][]color.Color{
		{color.White, color.Black},
		{color.White, color.White},
	}
	frame3WantColors := [][]color.Color{
		{color.White, color.White},
		{color.White, color.Black},
	}
	wantColors := [][][]color.Color{
		backgroundWantColors,
		frame1WantColors,
		frame2WantColors,
		frame3WantColors,
	}

	for i := 0; i < 4; i++ {
		if actual, want := gifHelper.Delay(), wantDelays[i]; actual != want {
			t.Errorf(`Frame %d; delay = %v, want %v`, i, actual, want)
		}

		wantImage := wantColors[i]

		for y, row := range wantImage {
			for x, c := range row {
				r, g, b, a := gifHelper.At(x, y).RGBA()
				wr, wg, wb, wa := c.RGBA()
				if r != wr || g != wg || b != wb || a != wa {
					t.Errorf(
						`Frame %d @ pixel %d, %d: RGBA channels = %v %v %v %v, want %v %v %v %v`,
						i, x, y,
						r, g, b, a,
						wr, wg, wb, wa,
					)
				}
			}
		}

		if err := gifHelper.NextFrame(); err != nil {
			t.Fatalf(`Frame %d: should be infinite loop, NextFrame = %v`, i, err)
		}
	}
}

// TestUnknownDisposal loads an animated GIF and checks that it combines when
// the disposal method is unknown.
func TestUnknownDisposal(t *testing.T) {
	f, err := os.Open(getResource("spinning-2x2.gif"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		panic(err)
	}
	g.Disposal[0] = 0xFF // NOTE Unknown disposal
	gifHelper, err := NewHelper(g)
	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}

	frame1WantColors := [][]color.Color{
		{color.Black, color.White},
		{color.White, color.White},
	}

	for y, row := range frame1WantColors {
		for x, c := range row {
			r, g, b, a := gifHelper.Frames[1].At(x, y).RGBA()
			wr, wg, wb, wa := c.RGBA()
			if r != wr || g != wg || b != wb || a != wa {
				t.Errorf(
					`Frame 1 @ (%d, %d): RGBA = %v %v %v %v, want %v %v %v %v`,
					x, y,
					r, g, b, a,
					wr, wg, wb, wa,
				)
			}
		}
	}
}

// TestReplace2x2 Tests that the frames of the replace directive are properly generated.
func TestReplace2x2(t *testing.T) {
	f, err := os.Open(getResource("replace-2x2.gif"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	gifHelper, err := HelperFromReader(f)
	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}

	if r, g, b, _ := gifHelper.Frames[1].At(0, 0).RGBA(); r != 0 || g != 0 || b != 0 {
		t.Errorf(`Frame 2 Color @ 0, 0 = rgb(%d,  %d, %d), want rgb(0, 0, 0)`, r, g, b)
	}

	for _, p := range []image.Point{{1, 0}, {0, 1}, {1, 1}} {
		if _, _, _, alpha := gifHelper.Frames[1].At(p.X, p.Y).RGBA(); alpha != 0 {
			t.Errorf(
				`Frame 2 Alpha @ %d, %d = %d (previous disposal %#02x), want 0`,
				p.X,
				p.Y,
				alpha,
				gifHelper.Frames[0].disposalMethod,
			)
		}
	}
	if r, g, b, _ := gifHelper.Frames[2].At(1, 0).RGBA(); r != 0 || g != 0 || b != 0 {
		t.Errorf(`Frame 3 Color @ 1, 0 = rgb(%d,  %d, %d), want rgb(0, 0, 0)`, r, g, b)
	}

	for _, p := range []image.Point{{0, 0}, {0, 1}, {1, 1}} {
		if _, _, _, alpha := gifHelper.Frames[2].At(p.X, p.Y).RGBA(); alpha != 0 {
			t.Errorf(
				`Frame 3 Alpha @ %d, %d = %d (previous disposal %#02x), want 0`,
				p.X,
				p.Y,
				alpha,
				gifHelper.Frames[1].disposalMethod,
			)
		}
	}
}

// TestAnimationNoLoop loads an animated GIF and checks that it does not loop.
func TestAnimationNoLoop(t *testing.T) {
	f, err := os.Open(getResource("spinning-2x2-noloop.gif"))
	if err != nil {
		panic(err)
	}
	gifHelper, err := HelperFromReader(f)
	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}
	if _, ok := gifHelper.loopCount.(noLoop); !ok {
		t.Errorf(`loopCount is %T, want noLoop`, gifHelper.loopCount)
	}

	for i := 0; i < 3; i++ {
		gifHelper.NextFrame()
	}
	if err := gifHelper.NextFrame(); err != ErrAnimationComplete {
		t.Errorf(`NextFrame = %v, want %v`, err, ErrAnimationComplete)
	}
}

// TestAnimation3Loop loads an animated GIF and checks that it loops 3 times.
func TestAnimation3Loop(t *testing.T) {
	f, err := os.Open(getResource("spinning-2x2-3loop.gif"))
	if err != nil {
		panic(err)
	}
	gifHelper, err := HelperFromReader(f)
	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}
	if loopCount, ok := gifHelper.loopCount.(*countLoop); !ok {
		t.Errorf(`loopCount is %T, want *countLoop`, gifHelper.loopCount)
	} else if count := loopCount.count; count != 3 {
		t.Errorf(`count = %v, want 3`, count)
	}

	expectedNextLoops := []error{nil, nil, nil, ErrAnimationComplete}
	for i, want := range expectedNextLoops {
		for i := 0; i < 3; i++ {
			gifHelper.NextFrame()
		}
		if err := gifHelper.NextFrame(); err != want {
			t.Errorf(`loop %d: NextFrame = %v, want %v`, i+1, err, want)
		}
	}
	if err := gifHelper.NextFrame(); err != ErrAnimationComplete {
		t.Errorf(`NextFrame = %v, want %v`, err, ErrAnimationComplete)
	}
}

type alwaysErrReader struct{}

func (r alwaysErrReader) Read([]byte) (int, error) {
	return 0, errors.New("mock")
}

func thisDirOrPanic() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("Couldn't get directory of test")
	}
	return filepath.Dir(file)
}

func getResource(resourceName string) string {
	dir := thisDirOrPanic()
	return filepath.Join(dir, "..", "..", "_resources", "tests", "pkg", "gif", resourceName)
}
