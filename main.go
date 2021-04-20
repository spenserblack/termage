package main

import (
	"image"
	_ "image/jpeg"
	"log"
	"os"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"golang.org/x/term"

	"github.com/spenserblack/termage/cmd"
)

func main() {
	cmd.Execute()

	reader, err := os.Open(cmd.ImageFile)
	if err != nil {
		log.Fatalf("Couldn't open image file: %v", err)
	}
	defer reader.Close()

	im, format, err := image.Decode(reader)
	_ = format
	if err != nil {
		log.Fatalf("Couldn't decode image: %v", err)
	}

	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Couldn't get terminal size: %v", err)
	}
	if err := ui.Init(); err != nil {
		log.Fatalf("Couldn't initialize UI: %v", err)
	}
	defer ui.Close()

	i := widgets.NewImage(im)
	i.Border = false
	i.SetRect(0, 0, width, height)

	ui.Render(i)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}
