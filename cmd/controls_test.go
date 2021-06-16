package cmd

import "testing"

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
