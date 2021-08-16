package draw

import (
	"image"

	"github.com/gdamore/tcell/v2"

	"github.com/spenserblack/termage/internal/conversion"
)

// TitleBarPixels is the height of the title bar in "pixels"
const TitleBarPixels int = 1

func Redraw(s tcell.Screen, title string, rgbRunes conversion.RGBRunes, center image.Point) {
	s.Clear()
	Title(s, title)
	Image(s, rgbRunes, center)
	s.Show()
}

// Title draws an image title to a screen.
func Title(s tcell.Screen, title string) {
	runes := []rune(title)
	width, _ := s.Size()
	center := width / 2
	runesStart := center - (len(runes) / 2)
	for i, r := range runes {
		s.SetContent(runesStart+i, 0, r, nil, tcell.StyleDefault)
	}
}

// Image draws an image to a screen.
//
// Center is the center of the image relative to the screen's center, with
// center = 0, 0 meaning that the image is perfectly centered in the screen.
func Image(s tcell.Screen, rgbRunes conversion.RGBRunes, center image.Point) {
	width, height := rgbRunes.Width(), rgbRunes.Height()
	screenWidth, screenHeight := s.Size()
	xOrigin := screenWidth / 2
	yOrigin := (screenHeight - TitleBarPixels) / 2
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if (yOrigin-height/2)+y+center.Y < TitleBarPixels {
				continue
			}
			rgbRune := rgbRunes.At(x, y)
			runeColor := tcell.FromImageColor(rgbRune)
			runeStyle := tcell.StyleDefault.Foreground(runeColor)
			s.SetContent(
				(xOrigin-width/2)+(x+center.X),
				(yOrigin-height/2)+(y+center.Y)+TitleBarPixels,
				rgbRune.Rune,
				nil,
				runeStyle,
			)
		}
	}
}
