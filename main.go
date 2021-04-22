package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/nfnt/resize"
	"golang.org/x/term"

	"github.com/spenserblack/termage/cmd"
	"github.com/spenserblack/termage/pkg/conversion"
)

func main() {
	cmd.Execute()

	reader, err := os.Open(cmd.ImageFile)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	i, format, err := image.Decode(reader)
	_ = format // TODO Display format in title

	if err != nil {
		log.Fatal(err)
	}

	i = resizeImageToTerm(i)
	rgbRunes := conversion.RGBRunesFromImage(i)

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}
	s.SetStyle(tcell.StyleDefault)

	drawImage := func() {
		s.Clear()
		width, height := rgbRunes.Width(), rgbRunes.Height()
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				rgbRune := rgbRunes.At(x, y)
				runeColor := tcell.NewRGBColor(
					int32(rgbRune.R),
					int32(rgbRune.G),
					int32(rgbRune.B),
				)
				runeStyle := tcell.StyleDefault.Foreground(runeColor)
				s.SetContent(x, y, rgbRune.Rune, nil, runeStyle)
			}
		}
		s.Show()
	}

	drawImage()

	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			s.Sync()
			drawImage()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape {
				s.Fini()
				os.Exit(0)
			}
		}
	}
}

var mainTerm = int(os.Stdin.Fd())

func termSize() (width, height int) {
	var err error
	width, height, err = term.GetSize(mainTerm)
	if err != nil {
		log.Fatalf("Couldn't get terminal size: %v", err)
	}
	return
}

func resizeImageToTerm(i image.Image) image.Image {
	width, height := termSize()
	if width < height {
		return resize.Resize(uint(width), 0, i, resize.NearestNeighbor)
	}
	return resize.Resize(0, uint(height), i, resize.NearestNeighbor)
}
