package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/nfnt/resize"

	"github.com/spenserblack/termage/cmd"
	"github.com/spenserblack/termage/internal/files"
	"github.com/spenserblack/termage/internal/helpers"
	"github.com/spenserblack/termage/pkg/conversion"
)

const titleBarPixels = 1

func main() {
	cmd.Execute()

	var browser files.FileBrowser
	var err error

	if len(cmd.ImageFiles) != 0 {
		browser = files.FileBrowser{Filenames: cmd.ImageFiles}
	} else {
		browser, err = files.NewFileBrowser(cmd.ImageFile)
	}

	if err != nil {
		log.Fatal(err)
	}
	var (
		reader        *os.File
		originalImage image.Image
		resizedImage  image.Image
		format        string
	)

	_ = format // TODO Display format in title

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}
	s.SetStyle(tcell.StyleDefault)

	loadImage := func() {
		reader.Close()
		reader, err := os.Open(browser.Current())
		if err != nil {
			log.Fatal(err)
		}

		originalImage, format, err = image.Decode(reader)

		if err != nil {
			log.Fatal(err)
		}
		resizedImage = resizeImageToTerm(originalImage, s)
	}

	drawImage := func() {
		rgbRunes := conversion.RGBRunesFromImage(resizedImage)
		s.Clear()
		width, height := rgbRunes.Width(), rgbRunes.Height()
		for x := 0; x < width; x++ {
			for y := titleBarPixels; y < height; y++ {
				rgbRune := rgbRunes.At(x, y)
				runeColor := tcell.NewRGBColor(
					// NOTE Takes 32-bit int, but requires range 0-255
					int32(helpers.BitshiftTo8Bit(rgbRune.R>>8)),
					int32(helpers.BitshiftTo8Bit(rgbRune.G>>8)),
					int32(helpers.BitshiftTo8Bit(rgbRune.B>>8)),
				)
				runeStyle := tcell.StyleDefault.Foreground(runeColor)
				s.SetContent(x, y, rgbRune.Rune, nil, runeStyle)
			}
		}
		s.Show()
	}

	loadImage()
	drawImage()

	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			s.Sync()
			drawImage()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				s.Fini()
				os.Exit(0)
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'n':
					browser.Forward()

					loadImage()
					drawImage()
				case 'N':
					browser.Back()

					loadImage()
					drawImage()
				}
			}
		}
	}
}

func resizeImageToTerm(i image.Image, s tcell.Screen) image.Image {
	width, height := s.Size()
	height -= titleBarPixels
	if width < height {
		return resize.Resize(uint(width), 0, i, resize.NearestNeighbor)
	}
	return resize.Resize(0, uint(height), i, resize.NearestNeighbor)
}
