package conversion

import (
	"image"
	"image/color"
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

// TestRGBRuneCreation tests that an RGBRune is created with the correct pixel
// placement and values.
func TestRGBRuneCreation(t *testing.T) {
	min := image.Point{0, 0}
	max := image.Point{2, 2}

	img := image.NewRGBA(image.Rectangle{min, max})

	red := color.RGBA{0xFF, 0, 0, 0xFF}
	green := color.RGBA{0, 0xFF, 0, 0xFF}
	blue := color.RGBA{0, 0, 0xFF, 0xFF}
	transparent := color.RGBA{0, 0, 0, 0}

	img.Set(0, 0, red)
	img.Set(1, 0, green)
	img.Set(0, 1, blue)
	img.Set(1, 1, transparent)

	rgbRunes := RGBRunesFromImage(img)

	if l, exp := len(rgbRunes.rgbRunes), 4; l != exp {
		t.Fatalf(`%v rgb runes (%v), want %v`, l, rgbRunes.rgbRunes, exp)
	}
	if width := rgbRunes.Width(); width != 2 {
		t.Fatalf(`width = %v, want 1`, width)
	}
	if height := rgbRunes.Height(); height != 2 {
		t.Fatalf(`width = %v, want 1`, height)
	}

	originalColors := [4]color.Color{red, green, blue, transparent}
	expectedReds := [4]uint32{0xFFFF, 0, 0, 0}
	expectedGreens := [4]uint32{0, 0xFFFF, 0, 0}
	expectedBlues := [4]uint32{0, 0, 0xFFFF, 0}
	expectedAlphas := [4]rune{'█', '█', '█', ' '}
	channelText := [3]string{"red", "green", "blue"}

	for i, oc := range originalColors {
		rgbRune := rgbRunes.At(i&0b01, i>>1)
		expectedChannels := [3]uint32{
			expectedReds[i],
			expectedGreens[i],
			expectedBlues[i],
		}
		actualChannels := [3]uint32{
			rgbRune.R,
			rgbRune.G,
			rgbRune.B,
		}

		for cIndex, ec := range expectedChannels {
			ac := actualChannels[cIndex]
			channelName := channelText[cIndex]
			if ac != ec {
				t.Errorf(`%s channel for %v = %v, want %v`, channelName, oc, ac, ec)
			}
		}
		if er, ar := expectedAlphas[i], rgbRune.Rune; ar != er {
			t.Errorf(`rune for %v = %q, want %q`, oc, ar, er)
		}
	}
}
