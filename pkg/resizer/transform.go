package resizer

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

func Transform(imgr io.ReadSeeker, w, h int, fit string) ([]byte, string, error) {
	img, format, err := NormaliseDecode(imgr)
	if err != nil {
		return nil, "", err
	}

	buf := new(bytes.Buffer)
	switch format {
	case "gif":
		transform := func(img image.Image) image.Image {
			return resize(img, w, h, fit)
		}
		err = processGIF(buf, imgr, transform)
		if err != nil {
			return nil, "", err
		}
	case "jpeg":
		img = resize(img, w, h, fit)

		options := jpeg.Options{Quality: 85}
		err = jpeg.Encode(buf, img, &options)
		if err != nil {
			return nil, "", err
		}
	case "png":
		img = resize(img, w, h, fit)

		err = png.Encode(buf, img)
		if err != nil {
			return nil, "", err
		}
	default:
		return nil, "", fmt.Errorf("unknown type")
	}

	return buf.Bytes(), format, nil
}
