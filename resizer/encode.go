package resizer

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

func Encode(w io.Writer, img image.Image, format string) error {
	switch format {
	case "jpeg":
		options := jpeg.Options{Quality: 85}
		return jpeg.Encode(w, img, &options)

	case "png":
		return png.Encode(w, img)

	case "gif":
		options := gif.Options{NumColors: 256}
		return gif.Encode(w, img, &options)

	default:
		return fmt.Errorf("unknown type")
	}
}
