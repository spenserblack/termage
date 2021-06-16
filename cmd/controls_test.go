package cmd

import (
	"bytes"
	"testing"
)

// TestControlString checks that a control has the proper string representation.
func TestControlString(t *testing.T) {
	expectedToControl := map[string]controlMapping{
		"n         Next image":       controlMapping{"n", "Next image"},
		"Esc       Exit application": controlMapping{"Esc", "Exit application"},
	}
	for want, control := range expectedToControl {
		if actual := control.String(); want != actual {
			t.Errorf(`%#v got %q, want %q`, control, actual, want)
		}
	}
}

// TestControlCommand makes sure that the correct subcommand is executed.
func TestControlCommand(t *testing.T) {
	out := new(bytes.Buffer)
	RootCmd.SetOut(out)
	RootCmd.SetArgs([]string{"controls"})

	command, err := RootCmd.ExecuteC()

	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}

	want := "controls"

	if actual := command.Name(); actual != want {
		t.Fatalf(`command name = %q, want %q`, actual, want)
	}
}
