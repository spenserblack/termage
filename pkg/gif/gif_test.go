package gif

import (
	"errors"
	"testing"
)

// TestInvalidReader checks that an error will be returned if the reader cannot
// be decoded to a GIF.
func TestInvalidReader(t *testing.T) {
	_, err := HelperFromReader(alwaysErrReader{})
	if err == nil {
		t.Errorf(`err = nil`)
	}
}

type alwaysErrReader struct{}

func (r alwaysErrReader) Read([]byte) (int, error) {
	return 0, errors.New("mock")
}
