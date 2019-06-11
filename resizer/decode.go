package resizer

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

func NormaliseDecode(r io.ReadSeeker) (image.Image, string, error) {
	img, format, err := image.Decode(r)
	if err != nil || format != "jpeg" {
		return img, format, err
	}

	if _, err := r.Seek(0, 0); err != nil {
		// Ignore seek error and just return what we got
		return img, format, nil
	}

	x, err := exif.Decode(r)
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
