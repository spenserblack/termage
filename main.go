package main

import (
	"log"

	"github.com/rivo/tview"

	"github.com/spenserblack/termage/cmd"
)

func main() {
	cmd.Execute()

	app := tview.NewApplication()
	imageView := tview.NewTextView().
		SetDynamicColors(true).
		SetText("Image will be drawn here.\nIt can be zoomed into and scrolled.").
		SetChangedFunc(func() {
			app.Draw()
		})

	if err := app.SetRoot(imageView, true).SetFocus(imageView).Run(); err != nil {
		log.Fatal(err)
	}
}
