package draw

import "github.com/gdamore/tcell/v2"

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
