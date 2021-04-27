package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"

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
		title         string
		// Modifiers for x and y coordinates of image
		xMod, yMod int
	)

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
		currentFile := browser.Current()
		reader, err := os.Open(currentFile)
		if err != nil {
			log.Fatal(err)
		}

		originalImage, format, err = image.Decode(reader)
		title = fmt.Sprintf("%v [%v]", filepath.Base(currentFile), format)

		if err != nil {
			log.Fatal(err)
		}
		resizedImage = resizeImageToTerm(originalImage, s)
	}

	drawTitle := func() {
		runes := []rune(title)
		width, _ := s.Size()
		center := width / 2
		runesStart := center - (len(runes) / 2)
		for i, r := range runes {
			s.SetContent(runesStart+i, 0, r, nil, tcell.StyleDefault)
		}
	}

	drawImage := func() {
		rgbRunes := conversion.RGBRunesFromImage(resizedImage)
		width, height := rgbRunes.Width(), rgbRunes.Height()
		screenWidth, screenHeight := s.Size()
		xOrigin := screenWidth / 2
		yOrigin := (screenHeight - titleBarPixels) / 2
		for x := 0; x < width; x++ {
			for y := titleBarPixels; y < height; y++ {
				if (yOrigin-height/2)+y+yMod <= titleBarPixels {
					continue
				}
				rgbRune := rgbRunes.At(x, y)
				runeColor := tcell.NewRGBColor(
					// NOTE Takes 32-bit int, but requires range 0-255
					int32(helpers.BitshiftTo8Bit(rgbRune.R>>8)),
					int32(helpers.BitshiftTo8Bit(rgbRune.G>>8)),
					int32(helpers.BitshiftTo8Bit(rgbRune.B>>8)),
				)
				runeStyle := tcell.StyleDefault.Foreground(runeColor)
				s.SetContent(
					(xOrigin-width/2)+(x+xMod),
					(yOrigin-height/2)+(y+yMod),
					rgbRune.Rune,
					nil,
					runeStyle,
				)
			}
		}
	}

	draw := func() {
		s.Clear()
		drawTitle()
		drawImage()
		s.Show()
	}

	shiftLeft := func(screenWidth, imageWidth int) {
		if xMod > (screenWidth-imageWidth)/2 {
			xMod--
		}
	}
	shiftRight := func(screenWidth, imageWidth int) {
		if xMod < (imageWidth-screenWidth)/2 {
			xMod++
		}
	}
	shiftUp := func(screenHeight, imageHeight int) {
		if yMod > (screenHeight-imageHeight)/2 {
			yMod--
		}
	}
	shiftDown := func(screenHeight, imageHeight int) {
		if yMod < (imageHeight-screenHeight)/2 {
			yMod++
		}
	}

	loadImage()
	draw()

	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			resizedImage = resizeImageToTerm(originalImage, s)
			draw()
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
					draw()
				case 'N':
					browser.Back()

					loadImage()
					draw()
				case 'z':
					resizedImage = zoomImage(
						originalImage,
						10,
						resizedImage.Bounds().Max,
					)
					draw()
				case 'Z':
					resizedImage = zoomImage(
						originalImage,
						-10,
						resizedImage.Bounds().Max,
					)
					draw()
				case 'f':
					xMod = 0
					yMod = 0
					resizedImage = resizeImageToTerm(originalImage, s)
					draw()
				case 'h':
					width, _ := s.Size()
					shiftLeft(width, resizedImage.Bounds().Max.X)
					draw()
				case 'H':
					width, _ := s.Size()
					rightBound := resizedImage.Bounds().Max.X
					for i := 0; i < rightBound/10; i++ {
						shiftLeft(width, rightBound)
					}
					draw()
				case 'j':
					_, height := s.Size()
					shiftDown(height, resizedImage.Bounds().Max.Y)
					draw()
				case 'J':
					_, height := s.Size()
					bounds := resizedImage.Bounds()
					for i := 0; i < bounds.Max.Y/10; i++ {
						shiftDown(height, bounds.Max.Y)
					}
					draw()
				case 'k':
					_, height := s.Size()
					shiftUp(height, resizedImage.Bounds().Max.Y)
					draw()
				case 'K':
					_, height := s.Size()
					bottomBound := resizedImage.Bounds().Max.Y
					for i := 0; i < bottomBound/10; i++ {
						shiftUp(height, bottomBound)
					}
					draw()
				case 'l':
					width, _ := s.Size()
					shiftRight(width, resizedImage.Bounds().Max.X)
					draw()
				case 'L':
					width, _ := s.Size()
					bounds := resizedImage.Bounds()
					for i := 0; i < bounds.Max.X/10; i++ {
						shiftRight(width, bounds.Max.X)
					}
					draw()
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

func zoomImage(
	original image.Image,
	percentage int,
	maxBound image.Point,
) image.Image {
	return resize.Resize(
		uint(maxBound.X+(maxBound.X*percentage/100)),
		uint(maxBound.Y+(maxBound.Y*percentage/100)),
		original,
		resize.NearestNeighbor,
	)
}
