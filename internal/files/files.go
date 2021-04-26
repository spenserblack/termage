package files

import (
	"fmt"
	"os"
	"path/filepath"
)

// TODO Check out answer to https://stackoverflow.com/questions/25959386/how-to-check-if-a-file-is-a-valid-image
// TODO Make FileBrowser a list of file*names* so that it doesn't eat up a massive amount of memory.

// FileBrowser is a tool to browse through files in a directory.
//
// It is a helper to browse through image files.
type FileBrowser struct {
	index     int
	filenames []string
}

// NewFileBrowser creates a new file browser from a string pointing to a file
// or directory. If it is a file, then that is the initial file selected by the
// returned FileBrowser. If it is a directory, then the index will start at 0.
func NewFileBrowser(filename string) (browser FileBrowser, err error) {
	var currentDir string
	absoluteFilename, err := filepath.Abs(filename)
	if err != nil {
		return browser, newFileBrowserError(filename, err)
	}

	if fileInfo, err := os.Stat(filename); err != nil {
		return browser, newFileBrowserError(filename, err)
	} else if fileInfo.IsDir() {
		currentDir = absoluteFilename
	} else {
		currentDir = filepath.Dir(absoluteFilename)
	}

	matches, err := filepath.Glob(filepath.Join(currentDir, "*"))
	if err != nil {
		return
	}
	currentFileStats, err := os.Stat(absoluteFilename)
	if err != nil {
		return
	}

	for i, fpath := range matches {
		absFpath, err := filepath.Abs(fpath)
		if err != nil {
			return browser, err
		}
		fileStats, err := os.Stat(absFpath)
		if err != nil {
			return browser, err
		}
		if fileStats.IsDir() {
			continue
		}
		if os.SameFile(currentFileStats, fileStats) {
			browser.index = i
		}
		browser.filenames = append(browser.filenames, absFpath)
	}
	return
}

// Forward moves forward one file.
func (browser *FileBrowser) Forward() {
	browser.index = (browser.index + 1) % len(browser.filenames)
}

// Back moves back one file.
func (browser *FileBrowser) Back() {
	if browser.index <= 0 {
		browser.index = len(browser.filenames) - 1
	} else {
		browser.index--
	}
}

// Current gets the current file.
func (browser *FileBrowser) Current() string {
	return browser.filenames[browser.index]
}

func newFileBrowserError(filename string, err error) error {
	return fmt.Errorf("Couldn't initialize file browser for %q: %w", filename, err)
}
