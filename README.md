# termage

[![CI](https://github.com/spenserblack/termage/actions/workflows/ci.yml/badge.svg)](https://github.com/spenserblack/termage/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/spenserblack/termage)](https://goreportcard.com/report/github.com/spenserblack/termage)

Browse images in the terminal *with support for animated GIFs :tada:*

## Installation

```bash
go get -u github.com/spenserblack/termage
```

## Usage

### Help

```bash
termage --help
```

### Browse all images in a directory

#### Starting from first image in directory

```bash
termage path/to/dir/
```

#### Starting from a specific image

```bash
termage path/to/dir/image
```

### Browse a specific subset of images

```bash
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
- `Esc`: Exit application (you may need to press twice due to an upstream issue)
