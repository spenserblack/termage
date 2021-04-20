package cmd

import (
	"errors"
	"fmt"
	"os"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/spf13/cobra"
)

var (
	imageFile string
	rootCmd   = &cobra.Command{
		Use:   "asciiimage <FILE | DIRECTORY>",
		Short: "Browse image files as ASCII in your terminal",
		Long: `This application is a tool to view your image files
as ASCII art without leaving your terminal`,
		Args: cobra.ExactArgs(1),
		RunE: rootRun,
	}
)

func rootRun(cmd *cobra.Command, args []string) error {
	fmt.Println("file: ", args[0])
	if err := ui.Init(); err != nil {
		return errors.New("Please run in an environment that can initialize a TUI")
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Text = fmt.Sprintf("File: %v", args[0])
	p.SetRect(0, 0, 25, 5)

	ui.Render(p)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
