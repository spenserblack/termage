package files

import (
	"os"
	"path/filepath"
)

// FileBrowser is a tool to browse through files in a directory.
//
// It is a helper to browse through image files.
type FileBrowser struct {
	index int
	files []*os.File
}

// NewFileBrowser creates a new file browser from a string pointing to a file
// or directory. If it is a file, then that is the initial file selected by the
// returned FileBrowser. If it is a directory, then the index will start at 0.
func NewFileBrowser(filename string) (browser FileBrowser, err error) {
	var currentDir *os.File

	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	fileStats, err := file.Stat()

	if err != nil {
		return
	} else if fileStats.IsDir() {
		currentDir = file
	} else {
		dirname := filepath.Dir(filename)
		currentDir, err = os.Open(dirname)
		if err != nil {
			return browser, err
		}
		defer currentDir.Close()
	}

	// TODO Use a sane maximum?
	dirnames, err := currentDir.Readdirnames(0)
	if err != nil {
		return
	}

	// NOTE This assumes that at least the majority of files are images
	browser.files = make([]*os.File, 0, len(dirnames))
	for i, name := range dirnames {
		f, err := os.Open(name)
		if err != nil {
			return browser, err
		}

		stats, err := f.Stat()
		if err != nil {
			return browser, err
		}

		if os.SameFile(stats, fileStats) {
			browser.index = i
		}
		// TODO Limit to only image files
		browser.files = append(browser.files, f)
	}
	return
}

// Close closes all files in the FileBrowser.
func (browser *FileBrowser) Close() {
	for _, f := range browser.files {
		f.Close()
	}
}

// Forward moves forward one file.
func (browser *FileBrowser) Forward() {
	browser.index = (browser.index + 1) % len(browser.files)
}

// Back moves back one file.
func (browser *FileBrowser) Back() {
	if browser.index <= 0 {
		browser.index = len(browser.files) - 1
	} else {
		browser.index--
	}
}

// Current gets the current file.
func (browser *FileBrowser) Current() *os.File {
	return browser.files[browser.index]
}
