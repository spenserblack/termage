package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/spenserblack/termage/cmd"
)

func main() {
	cmd.Execute()

	rootBox := tview.NewBox().
		SetBorder(true).
		SetTitle(cmd.ImageFile).
		SetBackgroundColor(tcell.ColorDefault)
	if err := tview.NewApplication().SetRoot(rootBox, true).Run(); err != nil {
		log.Fatal(err)
	}
}
