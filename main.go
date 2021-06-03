package main

import (
	_ "image/gif"  // registers GIFs
	_ "image/jpeg" // registers JPEGs
	_ "image/png"  // registers PNGs

	"github.com/spenserblack/termage/cmd"
)

// Supported is a map of file extensions that are supported.
var supported map[string]struct{}

// Version is the current version at build.
var version string

func main() {
	cmd.Supported = supported
	cmd.Version = version
	cmd.Execute()
}

func init() {
	supportedExtensions := []string{
		"jpeg",
		"jpg",
		"png",
		"gif",
	}
	supported = make(map[string]struct{})
	for _, v := range supportedExtensions {
		supported["."+v] = struct{}{}
	}
}
