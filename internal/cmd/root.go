package cmd

import (
	"image"
	"log"
	"os"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gdamore/tcell/v2"

	"github.com/spenserblack/termage/internal/conversion"
	"github.com/spenserblack/termage/internal/draw"
	"github.com/spenserblack/termage/internal/files"
	"github.com/spenserblack/termage/internal/utils"
	"github.com/spenserblack/termage/pkg/gif"
)

const (
	pixelHeight float32 = 2.15
)

// Screen is the main screen that will be initialized and drawn to.
var Screen tcell.Screen

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
		// Modifiers for x and y coordinates of image
		xMod, yMod int
		images     chan image.Image = make(chan image.Image, 1)
		titleChan  chan string      = make(chan string, 1)
		errChan    chan error       = make(chan error, 1)
		doRedraw   chan struct{}    = make(chan struct{}, 1)
		shiftImg   chan Shift       = make(chan Shift)
		resetImg   chan struct{}    = make(chan struct{})
		zoomIn     chan struct{}    = make(chan struct{})
		zoomOut    chan struct{}    = make(chan struct{})
	)

	Screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := Screen.Init(); err != nil {
		log.Fatal(err)
	}
	Screen.SetStyle(tcell.StyleDefault)

	loadImage := func() {
		m, title, err := utils.LoadImage(browser.Current())
		titleChan <- title
		if err != nil && err != utils.ErrNotAnimated {
			errChan <- err
			return
		}
		images <- m
	}

	go func() {
		loadImage()
		var (
			fitZoom, currentZoom        Zoom
			title                       string
			currentImage                image.Image
			stopAnimation               chan struct{}            = make(chan struct{}, 1)
			nextFrame                   chan conversion.RGBRunes = make(chan conversion.RGBRunes)
			zoomChan                    chan Zoom                = make(chan Zoom)
			rgbRunes                    conversion.RGBRunes
			currentWidth, currentHeight int
		)
		zoomGif := func() {
			zoomChan <- currentZoom
		}
		for {
			select {
			case currentImage = <-images:
				xMod = 0
				yMod = 0
				stopAnimation <- struct{}{}
				stopAnimation = make(chan struct{}, 1)
				nextFrame = make(chan conversion.RGBRunes)
				maxWidth, maxHeight := Screen.Size()
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
					zoomChan = make(chan Zoom, 1)
					go AnimateGif(g, nextFrame, stopAnimation, zoomChan)
					go zoomGif()
					continue
				}
				resizedImage := currentZoom.TransImage(currentImage)
				rgbRunes = conversion.RGBRunesFromImage(resizedImage)
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
				draw.Image(Screen, rgbRunes, image.Point{xMod, yMod})
			case title = <-titleChan:
				draw.Title(Screen, title)
			case err := <-errChan:
				Screen.Clear()
				draw.Error(Screen, err)
				Screen.Show()
			case <-doRedraw:
				draw.Image(Screen, rgbRunes, image.Point{xMod, yMod})
			case <-zoomIn:
				if currentZoom < fitZoom && fitZoom < currentZoom+10 {
					currentZoom = fitZoom
				} else {
					currentZoom += 10
				}
				if _, ok := currentImage.(*gif.Helper); ok {
					go zoomGif()
					continue
				}
				resizedImage := currentZoom.TransImage(currentImage)
				rgbRunes = conversion.RGBRunesFromImage(resizedImage)
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
				draw.Image(Screen, rgbRunes, image.Point{xMod, yMod})
			case <-zoomOut:
				if currentZoom < 11 {
					currentZoom = 1
				} else {
					currentZoom -= 10
				}
				xMod /= 10
				yMod /= 10
				if _, ok := currentImage.(*gif.Helper); ok {
					go zoomGif()
					continue
				}
				resizedImage := currentZoom.TransImage(currentImage)
				rgbRunes = conversion.RGBRunesFromImage(resizedImage)
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
				draw.Image(Screen, rgbRunes, image.Point{xMod, yMod})
			case <-resetImg:
				xMod = 0
				yMod = 0
				maxWidth, maxHeight := Screen.Size()
				maxWidth = int(float32(maxWidth) / pixelHeight)
				if currentImage == nil {
					break
				}
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
					go zoomGif()
					continue
				}
				resizedImage := currentZoom.TransImage(currentImage)
				rgbRunes = conversion.RGBRunesFromImage(resizedImage)
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
				draw.Redraw(Screen, title, rgbRunes, image.Point{xMod, yMod})
			case shift := <-shiftImg:
				width, height := Screen.Size()
				height -= draw.TitleBarPixels
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
				go draw.Image(Screen, rgbRunes, image.Point{xMod, yMod})
				rgbRunes = frame
				currentWidth, currentHeight = rgbRunes.Width(), rgbRunes.Height()
			}
		}
	}()
	for {
		switch ev := Screen.PollEvent().(type) {
		case *tcell.EventResize:
			resetImg <- struct{}{}
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				Screen.Fini()
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
					doRedraw <- struct{}{}
				case 'H':
					shiftImg <- Shift{image.Point{-10, 0}, true}
					doRedraw <- struct{}{}
				case 'j':
					shiftImg <- Shift{image.Point{0, 1}, false}
					doRedraw <- struct{}{}
				case 'J':
					shiftImg <- Shift{image.Point{0, 10}, true}
					doRedraw <- struct{}{}
				case 'k':
					shiftImg <- Shift{image.Point{0, -1}, false}
					doRedraw <- struct{}{}
				case 'K':
					shiftImg <- Shift{image.Point{0, -10}, true}
					doRedraw <- struct{}{}
				case 'l':
					shiftImg <- Shift{image.Point{1, 0}, false}
					doRedraw <- struct{}{}
				case 'L':
					shiftImg <- Shift{image.Point{10, 0}, true}
					doRedraw <- struct{}{}
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
	// NextFrameSem is used to let the animator know when the wait has completed
	// and the next frame is ready.
	nextFrameSem := make(chan error, 1)
	nextFrameSem <- nil
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
			if <-nextFrameSem != nil {
				return
			}
			nextFrame <- frames[index]
			go func() {
				time.Sleep(g.Delay())
				nextFrameSem <- g.NextFrame()
			}()
			index++
			if index >= max {
				index = 0
			}
		}
	}
}
