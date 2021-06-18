package cmd

import (
	"bytes"
	"os"
	"testing"

	internal "github.com/spenserblack/termage/internal/cmd"
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

// TestImageFiles checks that the root command will collect the positional
// arguments as image filepaths.
func TestImageFiles(t *testing.T) {
	args := []string{"path/to/image1.ext", "path/to/image2.example"}
	mainFunc = func([]string, map[string]struct{}) {}
	defer func() {
		mainFunc = internal.Root
	}()

	outErr := new(bytes.Buffer)
	RootCmd.SetErr(outErr)
	RootCmd.SetArgs(args)

	if _, err := RootCmd.ExecuteC(); err != nil {
		t.Fatalf(`err %v, want nil`, err)
	}

	for i, v := range args {
		if actual := ImageFiles[i]; actual != v {
			t.Errorf(`ImageFiles[%d] = %v, want %v`, i, actual, v)
		}
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
