package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	internal "github.com/spenserblack/termage/internal/cmd"
)

var (
	// ImageFiles contains the filepaths that the user has specified.
	// This will be used when user specifies more than 1 image.
	ImageFiles []string = nil
	rootCmd             = &cobra.Command{
		Use:   "termage {<FILE | DIRECTORY> | <FILES...>}",
		Short: "Browse image files as ASCII in your terminal",
		Long: `This application is a tool to view your image files
as ASCII art without leaving your terminal.
If a directory is passed, you will browse all images in that directory.
If a single image file is passed, you will browse all images in the same
directory as that image.
If multiple files are passed, then you will browse specifically those files.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ImageFiles = args
			internal.Root(ImageFiles)
		},
	}
)

// Execute runs this project's CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
