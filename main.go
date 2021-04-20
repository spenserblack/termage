package main

import (
	"fmt"
	"log"
	"os"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"golang.org/x/term"

	"github.com/spenserblack/termage/cmd"
)

func main() {
	cmd.Execute()

	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Couldn't get terminal size: %v", err)
	}
	if err := ui.Init(); err != nil {
		log.Fatalf("Couldn't initialize UI: %v", err)
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Border = false
	p.SetRect(0, 0, width, height)
	p.Text = fmt.Sprintf(
		"File: %v\nInner width: %v\nInner height: %v",
		cmd.ImageFile,
		p.Inner.Dx(),
		p.Inner.Dy(),
	)

	ui.Render(p)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}
