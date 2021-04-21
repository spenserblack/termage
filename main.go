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
	titleView := tview.NewTextView().
		SetText(cmd.ImageFile).
		SetTextAlign(tview.AlignCenter)
	footerView := tview.NewTextView().
		SetText("This is the footer").
		SetTextAlign(tview.AlignCenter)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(titleView, 0, 10, false).
		AddItem(imageView, 0, 80, false).
		AddItem(footerView, 0, 10, false)

	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		log.Fatal(err)
	}
}
