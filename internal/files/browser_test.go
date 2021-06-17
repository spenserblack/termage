package files

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var imageExtensions = map[string]struct{}{
	".jpeg": struct{}{},
	".jpg":  struct{}{},
	".png":  struct{}{},
	".gif":  struct{}{},
}

// TestBasicBrowserCreation checks that a file browser can be created with a
// directory where all the files are valid and should be read.
func TestBasicBrowserCreation(t *testing.T) {
	tempDir := t.TempDir()
	for _, ext := range []string{"jpg", "gif", "png", "JPEG"} {
		pattern := fmt.Sprintf("*.%s", ext)
		_, err := ioutil.TempFile(tempDir, pattern)
		if err != nil {
			panic(err)
		}
	}

	browser, err := NewFileBrowser(tempDir, imageExtensions)

	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}

	if actual, want := len(browser.Filenames), 4; actual != want {
		t.Errorf(`got %d files, want %d`, actual, want)
	}

	if i := browser.index; i != 0 {
		t.Errorf(`browser.index = %d, want 0`, i)
	}

	if browser.IsEmpty() {
		t.Errorf(`IsEmpty() = true, want false`)
	}
}

// TestSelectFile checks that the browser's index will be on a selected file
// if the filepath matches an existing file.
func TestSelectFile(t *testing.T) {
	tempDir := t.TempDir()
	for _, ext := range []string{"jpg", "gif", "png", "JPEG"} {
		pattern := fmt.Sprintf("*.%s", ext)
		_, err := ioutil.TempFile(tempDir, pattern)
		if err != nil {
			panic(err)
		}
	}
	tempFile, err := ioutil.TempFile(tempDir, "my-file-*.jpg")
	if err != nil {
		panic(err)
	}

	browser, err := NewFileBrowser(tempFile.Name(), imageExtensions)

	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}

	absTempFile, err := filepath.Abs(tempFile.Name())
	if err != nil {
		panic(err)
	}

	absCurrentFile, err := filepath.Abs(browser.Current())
	if err != nil {
		panic(err)
	}

	if absTempFile != absCurrentFile {
		t.Errorf(`%q is current file, want %q`, absCurrentFile, absTempFile)
	}
}

// TestSkipUnsupported checks that a file browser will not include
// unsupported extensions.
func TestSkipUnsupported(t *testing.T) {
	tempDir := t.TempDir()
	for _, ext := range []string{"jpg", "txt", "go"} {
		pattern := fmt.Sprintf("*.%s", ext)
		_, err := ioutil.TempFile(tempDir, pattern)
		if err != nil {
			panic(err)
		}
	}

	browser, err := NewFileBrowser(tempDir, imageExtensions)

	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}

	if actual, want := len(browser.Filenames), 1; actual != want {
		t.Errorf(`got %d files, want %d`, actual, want)
	}

	if actual, want := filepath.Ext(browser.Current()), ".jpg"; actual != want {
		t.Errorf(`file extension = %v, want %v`, actual, want)
	}
}

// TestSubdirSkipped checks that a file browser will not include
// subdirectories.
func TestSubdirSkipped(t *testing.T) {
	tempDir := t.TempDir()
	for _, ext := range []string{"jpg", "gif", "png"} {
		pattern := fmt.Sprintf("*.%s", ext)
		_, err := ioutil.TempFile(tempDir, pattern)
		if err != nil {
			panic(err)
		}
	}
	tempSubDir, err := ioutil.TempDir(tempDir, "temp-subdir")
	if err != nil {
		panic(err)
	}
	absTempSubDir, err := filepath.Abs(tempSubDir)
	if err != nil {
		panic(err)
	}

	browser, err := NewFileBrowser(tempDir, imageExtensions)

	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}

	if actual, want := len(browser.Filenames), 3; actual != want {
		t.Errorf(`got %d files, want %d`, actual, want)
	}

	for _, fpath := range browser.Filenames {
		absFpath, err := filepath.Abs(fpath)
		if err != nil {
			panic(err)
		}
		if absFpath == absTempSubDir {
			t.Errorf(`%q is a sub directory of %q`, absFpath, tempDir)
		}
	}
}

// TestEmptyDir checks that IsEmpty is true when an empty file is read.
func TestEmptyDir(t *testing.T) {
	tempDir := t.TempDir()

	browser, err := NewFileBrowser(tempDir, imageExtensions)

	if err != nil {
		t.Fatalf(`err = %v, want nil`, err)
	}

	if !browser.IsEmpty() {
		t.Errorf(`IsEmpty() = false, want true`)
	}
}

// TestBrowserForward ensures that the browser increments the index and wraps
// to max index.
func TestBrowserForward(t *testing.T) {
	fb := FileBrowser{1, []string{"1", "2", "3"}}

	for _, want := range []string{"3", "1"} {
		fb.Forward()
		if actual := fb.Current(); actual != want {
			t.Errorf(`Did not go to next index; got %q @ %d, want %q`, actual, fb.index, want)
		}
	}
}

// TestBrowserBack ensures that the browser decrements the index and wraps
// to max index.
func TestBrowserBack(t *testing.T) {
	fb := FileBrowser{1, []string{"1", "2", "3"}}

	for _, want := range []string{"1", "3"} {
		fb.Back()
		if actual := fb.Current(); actual != want {
			t.Errorf(`Did not go to next index; got %q @ %d, want %q`, actual, fb.index, want)
		}
	}
}

// TestNonexistent checks that an error is returned when a non-existent
// directory/file is attempted to be opened.
func TestNonexistent(t *testing.T) {
	_, err := NewFileBrowser("...", imageExtensions)

	if err == nil {
		t.Fatalf(`err = nil`)
	}
}

// TestOsStatError checks that an error is returned if os.Stat fails for
// any reason.
func TestOsStatError(t *testing.T) {
	mockError := errors.New("mocked")
	osStat = func(string) (os.FileInfo, error) {
		return nil, mockError
	}
	defer func() {
		osStat = os.Stat
	}()

	tempDir := t.TempDir()

	_, err := NewFileBrowser(tempDir, imageExtensions)
	want := fmt.Errorf("Couldn't initialize file browser for %q: %w", tempDir, mockError)

	if err.Error() != want.Error() {
		t.Errorf(`err = %v, want %v`, err, want)
	}
}
