# termage

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/spenserblack/termage)
[![CI](https://github.com/spenserblack/termage/actions/workflows/ci.yml/badge.svg)](https://github.com/spenserblack/termage/actions/workflows/ci.yml)
[![CodeQL](https://github.com/spenserblack/termage/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/spenserblack/termage/actions/workflows/github-code-scanning/codeql)
[![Go Report Card](https://goreportcard.com/badge/github.com/spenserblack/termage)](https://goreportcard.com/report/github.com/spenserblack/termage)
[![codecov](https://codecov.io/gh/spenserblack/termage/branch/master/graph/badge.svg)](https://codecov.io/gh/spenserblack/termage)

[![GitHub all releases](https://img.shields.io/github/downloads/spenserblack/termage/total)][latest-release]

Browse images in the terminal *with support for animated GIFs :tada:*

## Examples

### PNG image

![Viewing a PNG image](./_resources/viewing_png.png "PNG image")

### Animated GIF

![Viewing an animated GIF](./_resources/viewing_gif.gif "Animated GIF")

## Installation

The following commands simplify installing from the [latest release][latest-release].

### Linux/MacOS

```shell
curl https://raw.githubusercontent.com/spenserblack/termage/HEAD/install.sh | bash
```

### PowerShell

```powershell
Invoke-WebRequest "https://raw.githubusercontent.com/spenserblack/termage/HEAD/install.ps1" | Invoke-Expression
```

## Usage

### Help

```sh
termage --help
```

### Browse all images in a directory

#### Starting from first image in directory

```sh
termage path/to/dir/
```

#### Starting from a specific image

```sh
termage path/to/dir/image
```

### Browse a specific subset of images

```sh
termage path/to/image1 path/to/image2 # ...
```

## Controls

- `n`: Next image
- `N`: Previous image
- `z`: Increase zoom by 10 percentiles
- `Z`: Decrease zoom by 10 percentiles
- `f`: Fit to screen
- `h`: Scroll left one pixel
- `H`: Scroll left 10%
- `j`: Scroll down one pixel
- `J`: Scroll down 10%
- `k`: Scroll up one pixel
- `K`: Scroll up 10%
- `l`: Scroll right one pixel
- `L`: Scroll right 10%
- `Esc`: Exit application

## Supported Formats

- PNG
- JPEG
- GIF

[latest-release]: https://github.com/spenserblack/termage/releases/latest
