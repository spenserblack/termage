package conversion

import (
	"image"
	"image/color"
)

// AlphaChars contains runes representing alpha levels, from most transparent
// (lowest index) to most opaque (highest index).
var AlphaChars = [...]rune{
	' ',
	'░',
	'▒',
	'▓',
	'█',
}

// RGBRune represents a colored character.
type RGBRune struct {
	r, g, b uint32
	char    rune
}

// RGBRunes is a helper type for a slice of RGBRunes.
type RGBRunes struct {
	rgbRunes      []RGBRune
	width, height int
}

// RGBRuneFromColor converts a color into an RGBRune.
func RGBRuneFromColor(c color.Color) RGBRune {
	r, g, b, a := c.RGBA()

	return RGBRune{
		r,
		g,
		b,
		AlphaChars[int(a/(^uint32(0)/uint32(len(AlphaChars))))],
	}
}

// RGBRunesFromImage creates a slice of RGBRunes from an Image.
func RGBRunesFromImage(i image.Image) RGBRunes {
	bounds := i.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	rgbRunes := make([]RGBRune, 0, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgbRunes = append(rgbRunes, RGBRuneFromColor(i.At(x, y)))
		}
	}

	return RGBRunes{
		rgbRunes,
		width,
		height,
	}
}

// At gets the RGBRune at a point.
func (rgb *RGBRunes) At(x, y int) RGBRune {
	return rgb.rgbRunes[y*rgb.width+x]
}

// Width gets the width of the image as colored runes.
func (rgb *RGBRunes) Width() int {
	return rgb.width
}

// Height gets the height of the image as colored runes.
func (rgb *RGBRunes) Height() int {
	return rgb.height
}
