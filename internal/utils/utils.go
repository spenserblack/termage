package utils

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/spenserblack/termage/pkg/gif"
)

// ErrNotAnimated is a re-export signifying that a GIF could not be animated.
var ErrNotAnimated = gif.ErrNotAnimated

// Variables for that can be changed for testing.
var (
	open   = os.Open
	decode = image.Decode
)

// LoadImage returns the image data and the title of the image.
func LoadImage(filename string) (m image.Image, title string, err error) {
	reader, err := open(filename)
	if err != nil {
		err = fmt.Errorf("Couldn't open %q: %w", filename, err)
		return
	}
	defer reader.Close()

	m, format, err := decode(reader)
	if err != nil {
		err = fmt.Errorf("Couldn't decode %q: %w", filename, err)
		return
	}
	title = formatTitle(filename, format)
	if format == "gif" {
		reader.Seek(0, 0)
		var helper gif.Helper
		helper, err = gif.HelperFromReader(reader)
		if err != nil {
			return
		}
		m = &helper
		return
	}
	return
}

// FormatTitle creates a proper title for an image.
func formatTitle(filename, imageFormat string) string {
	return fmt.Sprintf("%v [%v]", filepath.Base(filename), imageFormat)
}
