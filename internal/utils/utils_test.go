package utils

import (
	"path/filepath"
	"testing"
)

// TestFormatTitle checks that a title would be created containing the name
// and the format.
func TestFormatTitle(t *testing.T) {
	fakeFilename := filepath.Join("..", "my-dir", "gopher.jpg")
	title := formatTitle(fakeFilename, "JPEG")
	want := "gopher.jpg [JPEG]"

	if title != want {
		t.Fatalf(`title = %q, want %q`, title, want)
	}
}
