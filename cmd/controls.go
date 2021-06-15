package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var controlText = []string{
	"n - Next image",
	"N - Previous image",
	"z - Increase zoom by 10 percentiles",
	"Z - Decrease zoom by 10 percentiles",
	"f - Fit to screen",
	"h - Scroll left one pixel",
	"H - Scroll left 10%",
	"j - Scroll down one pixel",
	"J - Scroll down 10%",
	"k - Scroll up one pixel",
	"K - Scroll up 10%",
	"l - Scroll right one pixel",
	"L - Scroll right 10%",
	"Esc - Exit application",
}

var controlsCmd = &cobra.Command{
	Use:   "controls",
	Short: "Print controls",
	Run: func(cmd *cobra.Command, args []string) {
		for _, message := range controlText {
			fmt.Println(message)
		}
	},
}

func init() {
	RootCmd.AddCommand(controlsCmd)
}
