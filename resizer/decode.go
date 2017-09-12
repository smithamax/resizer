package resizer

import (
	"bytes"
	"image"
	"io"
	"io/ioutil"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

func NormaliseDecode(r io.Reader) (image.Image, string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	br := bytes.NewReader(b)

	img, format, err := image.Decode(br)
	if err != nil || format != "jpeg" {
		return img, format, err
	}

	br.Seek(0, 0)

	x, err := exif.Decode(br)
	if err != nil {
		// Ignore exif error and just return what we got
		// TODO: Add logging of error
		return img, format, nil
	}

	orientation, err := x.Get(exif.Orientation)
	if err != nil {
		return img, format, nil
	}

	o, _ := orientation.Int(0)

	switch o {
	case 2:
		img = imaging.FlipH(img)
	case 3:
		img = imaging.Rotate180(img)
	case 4:
		img = imaging.FlipV(img)
	case 5:
		img = imaging.Transpose(img)
	case 6:
		img = imaging.Rotate90(img)
	case 7:
		img = imaging.Transverse(img)
	case 8:
		img = imaging.Rotate270(img)
	}

	return img, format, nil
}
