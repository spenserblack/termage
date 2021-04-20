package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// ImageFile is the filepath to the initial file/directory.
	ImageFile string
	rootCmd   = &cobra.Command{
		Use:   "termage <FILE | DIRECTORY>",
		Short: "Browse image files as ASCII in your terminal",
		Long: `This application is a tool to view your image files
as ASCII art without leaving your terminal`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ImageFile = args[0]
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
