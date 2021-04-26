package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// ImageFile is the filepath to the initial file/directory.
	// It will be empty if user passes more than 1 image.
	ImageFile string
	// ImageFiles contains the filepaths that the user has specified.
	// This will be used when user specifies more than 1 image.
	ImageFiles []string = nil
	rootCmd             = &cobra.Command{
		Use:   "termage <FILE | DIRECTORY>",
		Short: "Browse image files as ASCII in your terminal",
		Long: `This application is a tool to view your image files
as ASCII art without leaving your terminal.
If a directory is passed, you will browse all images in that directory.
If a single image file is passed, you will browse all images in the same
directory as that image.
If multiple files are passed, then you will browse specifically those files.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 1:
				ImageFile = args[0]
			default:
				ImageFiles = args
			}
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
