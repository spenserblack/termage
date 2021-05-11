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
