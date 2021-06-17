package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileBrowser is a tool to browse through files in a directory.
//
// It is a helper to browse through image files.
type FileBrowser struct {
	index     int
	Filenames []string
}

// OsStat is set to a variable so that it can be mocked.
var osStat = os.Stat

// NewFileBrowser creates a new file browser from a string pointing to a file
// or directory. If it is a file, then that is the initial file selected by the
// returned FileBrowser. If it is a directory, then the index will start at 0.
func NewFileBrowser(filename string, extensions map[string]struct{}) (browser FileBrowser, err error) {
	var currentDir string
	absoluteFilename, err := filepath.Abs(filename)
	if err != nil {
		return browser, newFileBrowserError(filename, err)
	}

	if fileInfo, err := osStat(filename); err != nil {
		return browser, newFileBrowserError(filename, err)
	} else if fileInfo.IsDir() {
		currentDir = absoluteFilename
	} else {
		currentDir = filepath.Dir(absoluteFilename)
	}

	matches, _ := filepath.Glob(filepath.Join(currentDir, "*"))
	currentFileStats, err := os.Stat(absoluteFilename)
	if err != nil {
		return
	}

	for _, fpath := range matches {
		// NOTE Skips if extension is not supported
		if _, ok := extensions[strings.ToLower(filepath.Ext(fpath))]; !ok {
			continue
		}
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
			browser.index = len(browser.Filenames)
		}
		browser.Filenames = append(browser.Filenames, absFpath)
	}
	return
}

// Forward moves forward one file.
func (browser *FileBrowser) Forward() {
	browser.index = (browser.index + 1) % len(browser.Filenames)
}

// Back moves back one file.
func (browser *FileBrowser) Back() {
	if browser.index <= 0 {
		browser.index = len(browser.Filenames) - 1
	} else {
		browser.index--
	}
}

// Current gets the current file.
func (browser *FileBrowser) Current() string {
	return browser.Filenames[browser.index]
}

func newFileBrowserError(filename string, err error) error {
	return fmt.Errorf("Couldn't initialize file browser for %q: %w", filename, err)
}

// IsEmpty checks if there are no files that can be browsed.
func (browser *FileBrowser) IsEmpty() bool {
	return len(browser.Filenames) == 0
}
