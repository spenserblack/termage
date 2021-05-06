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

// Shift is a wrapper around image.Point that specifies absolute shift
// or relative.
type Shift struct {
	image.Point
	relative bool
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
		shiftImg   chan Shift       = make(chan Shift)
		resetImg   chan struct{}    = make(chan struct{})
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
			case <-resetImg:
				xMod = 0
				yMod = 0
				resizedImage = resizeImageToTerm(i, s)
				draw()
			case shift := <-shiftImg:
				width, height := s.Size()
				height -= titleBarPixels
				bounds := resizedImage.Bounds()
				x, y := shift.X, shift.Y
				if shift.relative {
					x = x * bounds.Max.X / 100
					y = y * bounds.Max.Y / 100
				}
				if bounds.Max.X > width {
					xMod += x
					if xMod < (width-bounds.Max.X)/2 {
						xMod = (width - bounds.Max.X) / 2
					}
					if xMod > (bounds.Max.X-width)/2 {
						xMod = (bounds.Max.X - width) / 2
					}
				}
				if bounds.Max.Y > height {
					yMod += y
					if yMod < (height-bounds.Max.Y)/2 {
						yMod = (height - bounds.Max.Y) / 2
					}
					if yMod > (bounds.Max.Y-height)/2 {
						yMod = (bounds.Max.Y - height) / 2
					}
				}
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
			resetImg <- struct{}{}
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
					resetImg <- struct{}{}
				case 'h':
					shiftImg <- Shift{image.Point{-1, 0}, false}
					redraw <- struct{}{}
				case 'H':
					shiftImg <- Shift{image.Point{-10, 0}, true}
					redraw <- struct{}{}
				case 'j':
					shiftImg <- Shift{image.Point{0, 1}, false}
					redraw <- struct{}{}
				case 'J':
					shiftImg <- Shift{image.Point{0, 10}, true}
					redraw <- struct{}{}
				case 'k':
					shiftImg <- Shift{image.Point{0, -1}, false}
					redraw <- struct{}{}
				case 'K':
					shiftImg <- Shift{image.Point{0, -10}, true}
					redraw <- struct{}{}
				case 'l':
					shiftImg <- Shift{image.Point{1, 0}, false}
					redraw <- struct{}{}
				case 'L':
					shiftImg <- Shift{image.Point{10, 0}, true}
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
