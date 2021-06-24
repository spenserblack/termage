package gif

import (
	"errors"
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
	if actual, want := gifHelper.Delay(), 500*time.Millisecond; actual != want {
		t.Errorf(`Background 1: delay = %v, want %v`, actual, want)
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
