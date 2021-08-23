package cmd

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	internal "github.com/spenserblack/termage/internal/cmd"
)

// Vars for mocking.
var (
	osExit   = os.Exit
	mainFunc = internal.Root
)

var (
	// Supported is a map containing file extensions that should be
	// supported. Modify before command is executed to set the extensions
	// that should be supported.
	Supported map[string]struct{}
	// ImageFiles contains the filepaths that the user has specified.
	// This will be used when user specifies more than 1 image.
	ImageFiles []string = nil
	// RootCmd is the root cobra command that runs the image viewer.
	RootCmd = &cobra.Command{
		Use:   "termage {<FILE | DIRECTORY> | <FILES...>}",
		Short: "Browse image files as ASCII in your terminal",
		Long: heredoc.Doc(`
			This application is a tool to view your image files
			as ASCII art without leaving your terminal.
			If a directory is passed, you will browse all images in that directory.
			If a single image file is passed, you will browse all images in the same
			directory as that image.
			If multiple files are passed, then you will browse specifically those files.
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ImageFiles = args
			mainFunc(ImageFiles, Supported)
		},
		Version: "0.6.0",
	}
)

// Execute runs this project's CLI.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		errOut := RootCmd.ErrOrStderr()
		fmt.Fprintln(errOut, err)
		osExit(1)
	}
}
