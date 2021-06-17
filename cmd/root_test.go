package cmd

import (
	"bytes"
	"os"
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

// TestExecute checks that root.Execute would exit if an error is returned.
func TestExecute(t *testing.T) {
	// execute := RootCmd.Execute
	exited := false
	// RootCmd.Execute = func(*cobra.Command) error {
	// 	return errors.New("mock")
	// }
	osExit = func(int) {
		exited = true
	}

	defer func() {
		// RootCmd.Execute = execute
		osExit = os.Exit
	}()

	outErr := new(bytes.Buffer)
	RootCmd.SetErr(outErr)
	RootCmd.SetArgs([]string{"--option-that-will-never-be-used"})
	Execute()

	if !exited {
		t.Errorf(`Would not have exited on RootCmd execution error`)
	}
}
