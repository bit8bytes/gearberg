// Package qrcode generates QR code images for inventory items.
package qrcode

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	goqr "github.com/skip2/go-qrcode"
)

const size = 256
const labelHeight = 24
const labelBaseline = size + 17

// Generate returns a PNG-encoded QR code for the given content string with the
// label rendered below the code for easy visual identification.
func Generate(content, label string) ([]byte, error) {
	raw, err := goqr.Encode(content, goqr.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("qrcode.Generate: %w", err)
	}

	if label == "" {
		return raw, nil
	}

	qrImg, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("qrcode.Generate: decode png: %w", err)
	}

	out := image.NewRGBA(image.Rect(0, 0, size, size+labelHeight))
	draw.Draw(out, out.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(out, image.Rect(0, 0, size, size), qrImg, image.Point{}, draw.Src)

	textWidth := font.MeasureString(basicfont.Face7x13, label).Ceil()
	x := (size - textWidth) / 2
	(&font.Drawer{
		Dst:  out,
		Src:  image.NewUniform(color.Black),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, labelBaseline),
	}).DrawString(label)

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, fmt.Errorf("qrcode.Generate: encode png: %w", err)
	}
	return buf.Bytes(), nil
}
