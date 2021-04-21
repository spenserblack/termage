package conversion

import "image/color"

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
