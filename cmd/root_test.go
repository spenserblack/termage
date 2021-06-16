package cmd

import (
	"bytes"
	"testing"
)

// Test1ArgMinimum checks that the root command requires at least 1 argument.
func Test1ArgMinimum(t *testing.T) {
	outErr := new(bytes.Buffer)
	RootCmd.SetErr(outErr)
	RootCmd.SetArgs([]string{})

	_, err := RootCmd.ExecuteC()

	if err == nil {
		t.Fatalf(`err = nil`)
	}
}
