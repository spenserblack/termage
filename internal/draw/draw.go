package draw

import (
	"image"

	"github.com/gdamore/tcell/v2"
	"github.com/spenserblack/go-wordwrap"

	"github.com/spenserblack/termage/internal/conversion"
)

// TitleBarPixels is the height of the title bar in "pixels"
const TitleBarPixels int = 3

func Redraw(s tcell.Screen, title string, rgbRunes conversion.RGBRunes, center image.Point) {
	s.Clear()
	Title(s, title)
	Image(s, rgbRunes, center)
	s.Show()
}

// Title draws an image title to a screen.
func Title(s tcell.Screen, title string) {
	width, _ := s.Size()
	clearRow(s, 0, TitleBarPixels, width)
	center := width / 2
	lines := wordwrap.WordWrap(title, width)
	for row, line := range lines {
		runes := []rune(line)
		runesStart := center - (len(runes) / 2)
		for i, r := range runes {
			s.SetContent(runesStart+i, row, r, nil, tcell.StyleDefault)
		}
	}
	s.Show()
}

// Image draws an image to a screen.
//
// Center is the center of the image relative to the screen's center, with
// center = 0, 0 meaning that the image is perfectly centered in the screen.
func Image(s tcell.Screen, rgbRunes conversion.RGBRunes, center image.Point) {
	width, height := rgbRunes.Width(), rgbRunes.Height()
	screenWidth, screenHeight := s.Size()
	clearImage(s, screenWidth, screenHeight)
	xOrigin := screenWidth / 2
	yOrigin := (screenHeight - TitleBarPixels) / 2
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if (yOrigin-height/2)+y+center.Y < 0 {
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
	s.Show()
}

// Error draws an error to the screen.
//
// An error should be drawn if an image *cannot* be drawn. An error should not
// be drawn *over* an image.
func Error(s tcell.Screen, err error) {
	status := "cannot draw:"
	errStr := err.Error()
	width, height := s.Size()

	statusLines := wordwrap.WordWrap(status, width)
	errorLines := wordwrap.WordWrap(errStr, width)

	xOrigin := width / 2
	yOrigin := (height - TitleBarPixels) / 2

	statusY := TitleBarPixels + yOrigin - 1
	for i, line := range statusLines {
		statusStart := xOrigin - (len(line) / 2)
		for j, char := range []rune(line) {
			s.SetContent(statusStart+j, statusY+i, char, nil, tcell.StyleDefault)
		}
	}

	errorY := statusY + len(statusLines)
	for i, line := range errorLines {
		errStart := xOrigin - (len(line) / 2)
		for j, char := range []rune(line) {
			s.SetContent(errStart+j, errorY+i, char, nil, tcell.StyleDefault)
		}
	}
}

// ClearRow clears a single row of a Screen.
func ClearRow(s tcell.Screen, start, end int) {
	width, _ := s.Size()
	clearRow(s, start, end, width)
}

// clearRow is the inner function that takes the screen width as a parameter.
func clearRow(s tcell.Screen, start, end int, width int) {
	for row := start; row <= end; row++ {
		for cell := 0; cell < width; cell++ {
			s.SetContent(cell, row, ' ', nil, tcell.StyleDefault)
		}
	}
}

// ClearImage clears all rows where the image would be drawn.
func ClearImage(s tcell.Screen) {
	width, height := s.Size()
	clearImage(s, width, height)
}

// clearImage is the inner function that takes the screen size and height
// parameters.
func clearImage(s tcell.Screen, width, height int) {
	clearRow(s, TitleBarPixels, height, width)
}
