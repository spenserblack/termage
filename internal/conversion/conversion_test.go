package conversion

import (
	"testing"
)

type FakeColor uint32

func (c FakeColor) RGBA() (r, g, b, a uint32) {
	return 0, 0, 0, uint32(c)
}

// TestAlphaToRune checks that an alpha value would be converted to the
// expected rune when executing RGBRuneFromColor.
func TestAlphaToRune(t *testing.T) {
	for i, alpha := range [len(AlphaChars)]uint32{
		(0xFFFF / 5) - 1,
		2*(0xFFFF/5) - 1,
		3*(0xFFFF/5) - 1,
		4*(0xFFFF/5) - 1,
		0xFFFF - 1,
	} {
		expected := AlphaChars[i]
		actual := RGBRuneFromColor(FakeColor(alpha)).Rune

		if actual != expected {
			t.Errorf(`Alpha %v converted to %q, want %q`, alpha, actual, expected)
		}
	}
}

// TestRuneToRGBA checks that RGBRune correctly implements the image/color.Color
// interface.
func TestRuneToRGBA(t *testing.T) {
	transparent := RGBRune{0xFFFF, 0, 0, ' '}
	r, _, _, a := transparent.RGBA()
	if r != 0 {
		t.Errorf(`Transparent should have no color, red = %v`, r)
	}
	if a != 0 {
		t.Errorf(`' ' should result in completely transparent but got %v`, a)
	}

	opaqueRed := RGBRune{0xFFFF, 0, 0, '█'}
	r, g, b, a := opaqueRed.RGBA()
	if r != 0xFFFF {
		t.Errorf(`Opaque max red = %v, want %v`, r, 0xFFFF)
	}
	if g != 0 {
		t.Errorf(`Opaque 0 green = %v, want %v`, r, 0)
	}
	if b != 0 {
		t.Errorf(`Opaque 0 blue = %v, want %v`, r, 0)
	}
	if a != 0xFFFF {
		t.Errorf(`Opaque alpha = %v, want %v`, a, 0xFFFF)
	}
}
