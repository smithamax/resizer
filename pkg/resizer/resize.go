package resizer

import (
	"image"

	"github.com/disintegration/imaging"
)

var defaultFilter imaging.ResampleFilter = imaging.Lanczos

func resize(img image.Image, w, h int, fit string) image.Image {
	if w == 0 && h == 0 {
		return img
	}

	switch fit {
	case "clip", "":
		if w == 0 || h == 0 {
			return imaging.Resize(img, w, h, defaultFilter)
		}

		size := img.Bounds().Size()

		r := float64(w) / float64(h)
		ir := float64(size.X) / float64(size.Y)

		var nw, nh int
		if r > ir {
			nw = int(float64(h) * ir)
			nh = h
		} else {
			nw = w
			nh = int(float64(w) / ir)
		}
		return imaging.Resize(img, nw, nh, defaultFilter)

	case "crop":
		return imaging.Thumbnail(img, w, h, defaultFilter)

	case "max":
		size := img.Bounds().Size()
		return imaging.Fit(img, min(w, size.X), min(h, size.Y), defaultFilter)

	case "min":
		size := img.Bounds().Size()

		if size.X > w && size.Y > h {
			return imaging.Thumbnail(img, w, h, defaultFilter)
		}

		r := float64(w) / float64(h)
		ir := float64(size.X) / float64(size.Y)

		var nw, nh int
		if r > ir {
			nw = size.X
			nh = int(float64(size.X) / r)
		} else {
			nw = int(float64(size.Y) * r)
			nh = size.Y
		}
		return imaging.Thumbnail(img, nw, nh, defaultFilter)
	}
	return img
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
