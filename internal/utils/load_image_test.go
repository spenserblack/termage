package utils

import (
	"errors"
	"image"
	"image/color"
	_ "image/jpeg"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spenserblack/termage/pkg/gif"
)

var testError error = errors.New(":(")

// TestOpenImage checks that a JPG can be opened.
func TestOpenImage(t *testing.T) {
	m, title, err := LoadImage(getResource("pixel.jpg"))
	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}
	if want := "pixel.jpg [jpeg]"; title != want {
		t.Errorf(`title = %q, want %q`, title, want)
	}
	pixel := m.At(0, 0)
	r, g, b, a := pixel.RGBA()
	wantColor := color.White
	wr, wg, wb, wa := wantColor.RGBA()
	if r != wr || g != wg || b != wb || a != wa {
		t.Errorf(`pixel = %#v, want white (%#v)`, pixel, wantColor)
	}
}

// TestLoadNotAnimated checks that an unanimated GIF will be treated as such.
func TestLoadNotAnimated(t *testing.T) {
	m, _, err := LoadImage(getResource("not-animated-pixel.gif"))
	if err != ErrNotAnimated {
		t.Errorf(`err = %v, want %v`, err, ErrNotAnimated)
	}
	if _, ok := m.(*gif.Helper); ok {
		t.Errorf(`Image is an gif.Helper`)
	}
}

// TestLoadAnimated checks that an animated GIF will be treated as such.
func TestLoadAnimated(t *testing.T) {
	m, _, err := LoadImage(getResource("animated-pixel.gif"))
	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}
	if _, ok := m.(*gif.Helper); !ok {
		t.Errorf(`Image is %T, want gif.Helper`, m)
	}
}

// TestFailedOpenError checks that the error informs that the file couldn't be
// opened.
func TestFailedOpenError(t *testing.T) {
	mockOpen(testError)
	defer resetOpen()
	_, _, err := LoadImage("")
	if err == nil {
		t.Fatalf("err = nil")
	}
	if !strings.Contains(err.Error(), "Couldn't open") {
		t.Fatalf(`Unexpected error message: %q`, err)
	}
}

// TestFailedDecode checks that the error informs that the image couldn't be
// decoded.
func TestFailedDecode(t *testing.T) {
	mockOpen(nil)
	defer resetOpen()
	mockDecode(testError)
	defer resetDecode()
	_, _, err := LoadImage("")
	if err == nil {
		t.Fatalf("err = nil")
	}
	if !strings.Contains(err.Error(), "Couldn't decode") {
		t.Fatalf(`Unexpected error message: %q`, err)
	}
}

func mockOpen(err error) {
	open = func(string) (*os.File, error) {
		return nil, err
	}
}

func resetOpen() {
	open = os.Open
}

func mockDecode(err error) {
	decode = func(io.Reader) (image.Image, string, error) {
		return nil, "", err
	}
}

func resetDecode() {
	decode = image.Decode
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
	return filepath.Join(dir, "..", "..", "_resources", "tests", "internal", "utils", resourceName)
}
