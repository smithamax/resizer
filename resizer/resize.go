package resizer

import (
	"image"

	"github.com/disintegration/imaging"
)

var defaultFilter imaging.ResampleFilter = imaging.Lanczos

func Resize(img image.Image, w, h int, fit string) image.Image {
	if w == 0 && h == 0 {
		return img
	}

	switch fit {
	case "clip", "":
		if w == 0 || h == 0 {
			return imaging.Resize(img, w, h, defaultFilter)
		}
		return imaging.Fit(img, w, h, defaultFilter)

	case "crop":
		return imaging.Thumbnail(img, w, h, defaultFilter)

	case "max":
		size := img.Bounds().Size()
		return imaging.Fit(img, min(w, size.X), min(h, size.Y), defaultFilter)

	case "min":
		size := img.Bounds().Size()
		return imaging.Thumbnail(img, min(w, size.X), min(h, size.Y), defaultFilter)
	}
	return img
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
