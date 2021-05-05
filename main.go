package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/nfnt/resize"

	"github.com/spenserblack/termage/cmd"
	"github.com/spenserblack/termage/internal/files"
	"github.com/spenserblack/termage/internal/helpers"
	"github.com/spenserblack/termage/pkg/conversion"
	"github.com/spenserblack/termage/pkg/gif"
)

const titleBarPixels = 1

var supportedExtensions = []string{
	"jpeg",
	"jpg",
	"png",
	"gif",
}

// Image is a wrapper around image.Image with additional details.
type Image struct {
	image.Image
	title  string
	format string
}

type quit = struct{}

func main() {
	cmd.Execute()

	var browser files.FileBrowser
	var err error

	if len(cmd.ImageFiles) != 0 {
		browser = files.FileBrowser{Filenames: cmd.ImageFiles}
	} else {
		supported := make(map[string]struct{})
		for _, v := range supportedExtensions {
			supported["."+v] = struct{}{}
		}
		browser, err = files.NewFileBrowser(cmd.ImageFile, supported)
	}

	if err != nil {
		log.Fatal(err)
	}
	if browser.IsEmpty() {
		log.Fatalf("No valid images found in %q", cmd.ImageFile)
	}
	var (
		reader       *os.File
		resizedImage image.Image
		title        string
		// Modifiers for x and y coordinates of image
		xMod, yMod int
		images     chan Image = make(chan Image, 1)
		stop       chan quit
		redraw     chan struct{}    = make(chan struct{}, 1)
		resizeAbs  chan image.Point = make(chan image.Point) // resize bounds
		resizeRel  chan int         = make(chan int)         // percentage
	)

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}
	s.SetStyle(tcell.StyleDefault)

	// Can use a quit channel in case it's an animation
	loadImage := func() chan quit {
		stop := make(chan quit, 1)
		currentFile := browser.Current()
		err := error(nil)
		reader, err = os.Open(currentFile)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()

		originalImage, format, err := image.Decode(reader)
		title = fmt.Sprintf("%v [%v]", filepath.Base(currentFile), format)

		if err != nil {
			log.Fatal(err)
		}
		images <- Image{originalImage, title, format}
		if format == "gif" {
			reader.Seek(0, 0)
			gifHelper, err := gif.HelperFromReader(reader)
			if err != nil {
				return stop
			}
			go func() {
				for {
					select {
					case <-stop:
						return
					default:
						images <- Image{
							gifHelper.Current,
							title,
							format,
						}
						time.Sleep(gifHelper.Delay())
					}
					if err := gifHelper.NextFrame(); err != nil {
						return
					}
				}
			}()
		}
		return stop
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

	stop = loadImage()

	go func() {
		i := <-images
		for {
			select {
			case <-redraw:
				draw()
			case p := <-resizeAbs:
				resizedImage = resize.Resize(
					uint(p.X),
					uint(p.Y),
					i,
					resize.NearestNeighbor,
				)
				draw()
			case percent := <-resizeRel:
				bounds := resizedImage.Bounds()
				resizedImage = resize.Resize(
					uint(bounds.Max.X+(bounds.Max.X*percent/100)),
					uint(bounds.Max.Y+(bounds.Max.Y*percent/100)),
					i,
					resize.NearestNeighbor,
				)
				draw()
			case newImage := <-images:
				i = newImage
				resizedImage = resizeImageToTerm(i, s)
				title = i.title
				draw()
			}
		}
	}()
	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			width, height := s.Size()
			if width < height {
				height = 0
			} else {
				width = 0
			}
			resizeAbs <- image.Point{width, height - titleBarPixels}
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				s.Fini()
				os.Exit(0)
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'n':
					stop <- quit{}
					browser.Forward()

					stop = loadImage()
				case 'N':
					stop <- quit{}
					browser.Back()

					stop = loadImage()
				case 'z':
					resizeRel <- 10
				case 'Z':
					resizeRel <- -10
				case 'f':
					xMod = 0
					yMod = 0
					width, height := s.Size()
					if width < height {
						height = 0
					} else {
						width = 0
					}
					resizeAbs <- image.Point{width, height - titleBarPixels}
				case 'h':
					width, _ := s.Size()
					shiftLeft(width, resizedImage.Bounds().Max.X)
					redraw <- struct{}{}
				case 'H':
					width, _ := s.Size()
					rightBound := resizedImage.Bounds().Max.X
					for i := 0; i < rightBound/10; i++ {
						shiftLeft(width, rightBound)
					}
					redraw <- struct{}{}
				case 'j':
					_, height := s.Size()
					shiftDown(height, resizedImage.Bounds().Max.Y)
					redraw <- struct{}{}
				case 'J':
					_, height := s.Size()
					bounds := resizedImage.Bounds()
					for i := 0; i < bounds.Max.Y/10; i++ {
						shiftDown(height, bounds.Max.Y)
					}
					redraw <- struct{}{}
				case 'k':
					_, height := s.Size()
					shiftUp(height, resizedImage.Bounds().Max.Y)
					redraw <- struct{}{}
				case 'K':
					_, height := s.Size()
					bottomBound := resizedImage.Bounds().Max.Y
					for i := 0; i < bottomBound/10; i++ {
						shiftUp(height, bottomBound)
					}
					redraw <- struct{}{}
				case 'l':
					width, _ := s.Size()
					shiftRight(width, resizedImage.Bounds().Max.X)
					redraw <- struct{}{}
				case 'L':
					width, _ := s.Size()
					bounds := resizedImage.Bounds()
					for i := 0; i < bounds.Max.X/10; i++ {
						shiftRight(width, bounds.Max.X)
					}
					redraw <- struct{}{}
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
