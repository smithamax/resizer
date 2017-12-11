package resizer

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"willnorris.com/go/gifresize"
)

func Transform(imgb []byte, w, h int, fit string) ([]byte, string, error) {
	img, format, err := NormaliseDecode(bytes.NewReader(imgb))
	if err != nil {
		return nil, "", err
	}

	buf := new(bytes.Buffer)
	switch format {
	case "gif":
		transform := func(img image.Image) image.Image {
			return resize(img, w, h, fit)
		}
		err = gifresize.Process(buf, bytes.NewReader(imgb), transform)
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
