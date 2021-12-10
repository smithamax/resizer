// Copyright 2013 Daniel Pupius. All rights reserved.
// Copyright 2015 Google. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resizer

// Code in this file is adapted from willnorris.com/go/gifresize

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
)

// transformFunc is a function that transforms an image.
type transformFunc func(image.Image) image.Image

// Process the GIF read from r, applying transform to each frame, and writing
// the result to w.
func processGIF(w io.Writer, src []byte, transform transformFunc) error {
	r := bytes.NewReader(src)
	if transform == nil {
		_, err := io.Copy(w, r)
		return err
	}

	// Decode the original gif.
	im, err := gif.DecodeAll(r)
	if err != nil {
		return err
	}

	// HACK: If animated skip resize
	if len(im.Image) > 1 {
		_, err := w.Write(src)
		return err
	}

	// Create a new RGBA image to hold the incremental frames.
	firstFrame := im.Image[0].Bounds()
	b := image.Rect(0, 0, firstFrame.Dx(), firstFrame.Dy())
	img := image.NewRGBA(b)

	// Resize each frame.
	for index, frame := range im.Image {
		fmt.Println(index)
		bounds := frame.Bounds()
		previous := img
		draw.Draw(img, bounds, frame, bounds.Min, draw.Over)
		im.Image[index] = imageToPaletted(transform(img), frame.Palette)

		switch im.Disposal[index] {
		case gif.DisposalBackground:
			// I'm just assuming that the gif package will apply the appropriate
			// background here, since there doesn't seem to be an easy way to
			// access the global color table
			img = image.NewRGBA(b)
		case gif.DisposalPrevious:
			img = previous
		}
	}

	// Set image.Config to new height and width
	im.Config.Width = im.Image[0].Bounds().Max.X
	im.Config.Height = im.Image[0].Bounds().Max.Y

	return gif.EncodeAll(w, im)
}

func imageToPaletted(img image.Image, p color.Palette) *image.Paletted {
	b := img.Bounds()
	pm := image.NewPaletted(b, p)
	draw.FloydSteinberg.Draw(pm, b, img, image.ZP)
	return pm
}
