package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

type controlMapping struct {
	key    string
	action string
}

func (m controlMapping) String() string {
	return fmt.Sprintf("%-10s%s", m.key, m.action)
}

var controls = []controlMapping{
	controlMapping{"n", "Next image"},
	controlMapping{"N", "Previous image"},
	controlMapping{"z", "Increase zoom by 10 percentiles"},
	controlMapping{"Z", "Decrease zoom by 10 percentiles"},
	controlMapping{"f", "Fit to screen"},
	controlMapping{"h", "Scroll left one pixel"},
	controlMapping{"H", "Scroll left 10%"},
	controlMapping{"j", "Scroll down one pixel"},
	controlMapping{"J", "Scroll down 10%"},
	controlMapping{"k", "Scroll up one pixel"},
	controlMapping{"K", "Scroll up 10%"},
	controlMapping{"l", "Scroll right one pixel"},
	controlMapping{"L", "Scroll right 10%"},
	controlMapping{"Esc", "Exit application"},
}

var controlsCmd = &cobra.Command{
	Use:   "controls",
	Short: "Print controls",
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		for _, control := range controls {
			io.WriteString(out, fmt.Sprintf("%s\n", control))
		}
	},
}

func init() {
	RootCmd.AddCommand(controlsCmd)
}
