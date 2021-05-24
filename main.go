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
	"github.com/spenserblack/termage/pkg/conversion"
	"github.com/spenserblack/termage/pkg/gif"
)

const (
	titleBarPixels         = 1
	pixelHeight    float32 = 2.15
)

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

// Zoom is used to manage the zoom-level as a percentage.
type Zoom uint

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
		zoom       Zoom             = 100
		images     chan Image       = make(chan Image, 1)
		frames     chan image.Image = make(chan image.Image)
		stop       chan quit
		redraw     chan struct{} = make(chan struct{}, 1)
		shiftImg   chan Shift    = make(chan Shift)
		resetImg   chan struct{} = make(chan struct{})
		zoomIn     chan struct{} = make(chan struct{})
		zoomOut    chan struct{} = make(chan struct{})
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
		var err error
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
		if format == "gif" {
			reader.Seek(0, 0)
			gifHelper, err := gif.HelperFromReader(reader)
			if err != nil {
				images <- Image{originalImage, title, format}
				return stop
			}
			images <- Image{&gifHelper, title, format}
			go func() {
				for {
					select {
					case <-stop:
						return
					default:
						frames <- gifHelper.CurrentImage()
						time.Sleep(gifHelper.Delay())
					}
					if err := gifHelper.NextFrame(); err != nil {
						return
					}
				}
			}()
		} else {
			images <- Image{originalImage, title, format}
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
				runeColor := tcell.FromImageColor(rgbRune)
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
		var fitZoom Zoom
		for {
			select {
			case <-redraw:
				draw()
			case <-zoomIn:
				if zoom < fitZoom && fitZoom < zoom+10 {
					zoom = fitZoom
				} else {
					zoom += 10
				}
				resizedImage = zoom.TransImage(i)
				draw()
			case <-zoomOut:
				if zoom < 11 {
					zoom = 1
				} else {
					zoom -= 10
				}
				xMod /= 10
				yMod /= 10
				resizedImage = zoom.TransImage(i)
				draw()
			case <-resetImg:
				xMod = 0
				yMod = 0
				maxWidth, maxHeight := s.Size()
				maxWidth = int(float32(maxWidth) / pixelHeight)
				if maxWidth < maxHeight {
					zoom = Zoom(maxWidth * 100 / i.Bounds().Max.X)
				} else {
					zoom = Zoom(maxHeight * 100 / i.Bounds().Max.Y)
				}
				if zoom > 100 {
					zoom = 100
				}
				fitZoom = zoom
				resizedImage = zoom.TransImage(i)
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
			case nextFrame := <-frames:
				resizedImage = zoom.TransImage(nextFrame)
				draw()
			case newImage := <-images:
				i = newImage
				maxWidth, maxHeight := s.Size()
				maxWidth = int(float32(maxWidth) / pixelHeight)
				if maxWidth < maxHeight {
					zoom = Zoom(uint(maxWidth) * 100 / uint(i.Bounds().Max.X))
				} else {
					zoom = Zoom(uint(maxHeight) * 100 / uint(i.Bounds().Max.Y))
				}
				if zoom > 100 {
					zoom = 100
				}
				fitZoom = zoom
				resizedImage = zoom.TransImage(i)
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
					zoomIn <- struct{}{}
				case 'Z':
					zoomOut <- struct{}{}
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

// TransImage transforms an image by a zoom percentage.
func (percentage Zoom) TransImage(i image.Image) image.Image {
	bounds := i.Bounds()
	// NOTE Adjusts width of "pixels" to match height
	width := float32(bounds.Max.X) * pixelHeight
	return resize.Resize(
		uint(width)*uint(percentage)/100,
		uint(bounds.Max.Y)*uint(percentage)/100,
		i,
		resize.NearestNeighbor,
	)
}
