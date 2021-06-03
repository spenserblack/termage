package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the version that will be printed when running the "version"
// subcommand.
//
// It can be set at build-time with ldflags.
//  go build -ldflags "-X github.com/spenserblack/termage/cmd.Version=x.y.z" main.go
var Version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version of termage",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	if Version == "" {
		Version = "development"
	}
}
