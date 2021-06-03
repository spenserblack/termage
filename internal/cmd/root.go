package cmd

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gdamore/tcell/v2"

	"github.com/spenserblack/termage/internal/conversion"
	"github.com/spenserblack/termage/internal/files"
	"github.com/spenserblack/termage/pkg/gif"
)

const (
	titleBarPixels         = 1
	pixelHeight    float32 = 2.15
)

// Shift is a wrapper around image.Point that specifies absolute shift
// or relative.
type Shift struct {
	image.Point
	relative bool
}

// Zoom is used to manage the zoom-level as a percentage.
type Zoom uint

// Root is the main function to be run by the root command.
func Root(imageFiles []string, supported map[string]struct{}) {
	var browser files.FileBrowser
	var err error

	if len(imageFiles) == 1 {
		browser, err = files.NewFileBrowser(imageFiles[0], supported)
	} else {
		browser = files.FileBrowser{Filenames: imageFiles}
	}

	if err != nil {
		log.Fatal(err)
	}
	if browser.IsEmpty() {
		log.Fatalf("No valid images found in %q", imageFiles[0])
	}
	var (
		reader *os.File
		// Modifiers for x and y coordinates of image
		xMod, yMod int
		images     chan image.Image = make(chan image.Image, 1)
		titleChan  chan string      = make(chan string, 1)
		redraw     chan struct{}    = make(chan struct{}, 1)
		shiftImg   chan Shift       = make(chan Shift)
		resetImg   chan struct{}    = make(chan struct{})
		zoomIn     chan struct{}    = make(chan struct{})
		zoomOut    chan struct{}    = make(chan struct{})
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
		currentFile := browser.Current()
		var err error
		reader, err = os.Open(currentFile)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()

		originalImage, format, err := image.Decode(reader)
		titleChan <- fmt.Sprintf("%v [%v]", filepath.Base(currentFile), format)

		if err != nil {
			log.Fatal(err)
		}
		if format == "gif" {
			reader.Seek(0, 0)
			gifHelper, err := gif.HelperFromReader(reader)
			if err != nil {
				images <- originalImage
				return
			}
			images <- &gifHelper
		} else {
			images <- originalImage
		}
	}

	drawTitle := func(title string) {
		runes := []rune(title)
		width, _ := s.Size()
		center := width / 2
		runesStart := center - (len(runes) / 2)
		for i, r := range runes {
			s.SetContent(runesStart+i, 0, r, nil, tcell.StyleDefault)
		}
	}

	drawImage := func(rgbRunes conversion.RGBRunes) {
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

	draw := func(title string, rgbRunes conversion.RGBRunes) {
		s.Clear()
		drawTitle(title)
		drawImage(rgbRunes)
		s.Show()
	}

	go func() {
		loadImage()
		var fitZoom, currentZoom Zoom
		title := <-titleChan
		currentImage := <-images
		stopAnimation := make(chan struct{}, 1)
		var nextFrame chan conversion.RGBRunes
		var zoomChan chan Zoom
		var rgbRunes conversion.RGBRunes
		var currentWidth, currentHeight int
		maxWidth, maxHeight := s.Size()
		maxWidth = int(float32(maxWidth) / pixelHeight)
		if maxWidth < maxHeight {
			currentZoom = Zoom(uint(maxWidth) * 100 / uint(currentImage.Bounds().Max.X))
		} else {
			currentZoom = Zoom(uint(maxHeight) * 100 / uint(currentImage.Bounds().Max.Y))
		}
		if currentZoom > 100 {
			currentZoom = 100
		}
		fitZoom = currentZoom
		if g, ok := currentImage.(*gif.Helper); ok {
			nextFrame = make(chan conversion.RGBRunes)
			zoomChan = make(chan Zoom)
			go AnimateGif(g, nextFrame, stopAnimation, zoomChan)
			zoomChan <- currentZoom
		}
		for {
			select {
			case currentImage = <-images:
				stopAnimation <- struct{}{}
				stopAnimation = make(chan struct{}, 1)
				maxWidth, maxHeight := s.Size()
				maxWidth = int(float32(maxWidth) / pixelHeight)
				if maxWidth < maxHeight {
					currentZoom = Zoom(uint(maxWidth) * 100 / uint(currentImage.Bounds().Max.X))
				} else {
					currentZoom = Zoom(uint(maxHeight) * 100 / uint(currentImage.Bounds().Max.Y))
				}
				if currentZoom > 100 {
					currentZoom = 100
				}
				fitZoom = currentZoom
				if g, ok := currentImage.(*gif.Helper); ok {
					nextFrame = make(chan conversion.RGBRunes)
					zoomChan = make(chan Zoom)
					go AnimateGif(g, nextFrame, stopAnimation, zoomChan)
					zoomChan <- currentZoom
					continue
				}
				resizedImage := currentZoom.TransImage(currentImage)
				rgbRunes = conversion.RGBRunesFromImage(resizedImage)
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
				draw(title, rgbRunes)
			case title = <-titleChan:
			case <-redraw:
				draw(title, rgbRunes)
			case <-zoomIn:
				if currentZoom < fitZoom && fitZoom < currentZoom+10 {
					currentZoom = fitZoom
				} else {
					currentZoom += 10
				}
				if _, ok := currentImage.(*gif.Helper); ok {
					zoomChan <- currentZoom
					continue
				}
				resizedImage := currentZoom.TransImage(currentImage)
				rgbRunes = conversion.RGBRunesFromImage(resizedImage)
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
				draw(title, rgbRunes)
			case <-zoomOut:
				if currentZoom < 11 {
					currentZoom = 1
				} else {
					currentZoom -= 10
				}
				xMod /= 10
				yMod /= 10
				if _, ok := currentImage.(*gif.Helper); ok {
					zoomChan <- currentZoom
					continue
				}
				resizedImage := currentZoom.TransImage(currentImage)
				rgbRunes = conversion.RGBRunesFromImage(resizedImage)
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
				draw(title, rgbRunes)
			case <-resetImg:
				xMod = 0
				yMod = 0
				maxWidth, maxHeight := s.Size()
				maxWidth = int(float32(maxWidth) / pixelHeight)
				if maxWidth < maxHeight {
					currentZoom = Zoom(maxWidth * 100 / currentImage.Bounds().Max.X)
				} else {
					currentZoom = Zoom(maxHeight * 100 / currentImage.Bounds().Max.Y)
				}
				if currentZoom > 100 {
					currentZoom = 100
				}
				fitZoom = currentZoom
				if _, ok := currentImage.(*gif.Helper); ok {
					zoomChan <- currentZoom
					continue
				}
				resizedImage := currentZoom.TransImage(currentImage)
				rgbRunes = conversion.RGBRunesFromImage(resizedImage)
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
				draw(title, rgbRunes)
			case shift := <-shiftImg:
				width, height := s.Size()
				height -= titleBarPixels
				x, y := shift.X, shift.Y
				if shift.relative {
					x = x * currentWidth / 100
					y = y * currentHeight / 100
				}
				if currentWidth > width {
					xMod += x
					if xMod < (width-currentWidth)/2 {
						xMod = (width - currentWidth) / 2
					}
					if xMod > (currentWidth-width)/2 {
						xMod = (currentWidth - width) / 2
					}
				}
				if currentHeight > height {
					yMod += y
					if yMod < (height-currentHeight)/2 {
						yMod = (height - currentHeight) / 2
					}
					if yMod > (currentHeight-height)/2 {
						yMod = (currentHeight - height) / 2
					}
				}
			case frame := <-nextFrame:
				draw(title, frame)
				rgbRunes = frame
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
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
					browser.Forward()

					loadImage()
				case 'N':
					browser.Back()

					loadImage()
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
	return imaging.Resize(
		i,
		int(width)*int(percentage)/100,
		bounds.Max.Y*int(percentage)/100,
		imaging.Linear,
	)
}

// AnimateGif is a helper to fire off animation events at the correct time.
func AnimateGif(g *gif.Helper, nextFrame chan conversion.RGBRunes, stop chan struct{}, zoomChan chan Zoom) {
	index := 0
	max := len(g.Frames)
	frames := make([]conversion.RGBRunes, max, max)
	zoom := <-zoomChan
	for i, v := range g.Frames {
		zoomedImage := zoom.TransImage(v)
		frames[i] = conversion.RGBRunesFromImage(zoomedImage)
	}
	for {
		select {
		case <-stop:
			return
		case zoom = <-zoomChan:
			for i, v := range g.Frames {
				zoomedImage := zoom.TransImage(v)
				frames[i] = conversion.RGBRunesFromImage(zoomedImage)
			}
		default:
			nextFrame <- frames[index]
			time.Sleep(g.Delay())
			if err := g.NextFrame(); err != nil {
				return
			}
			index++
			if index >= max {
				index = 0
			}
		}
	}
}
