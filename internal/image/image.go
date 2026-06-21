// Package image decodes JPEG, PNG, and WebP uploads, resizes them to a safe
// maximum dimension, and re-encodes them as JPEG or PNG. Re-encoding strips any
// malicious payloads that could be embedded in the original file (polyglots,
// embedded scripts, etc.).
package image

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"

	_ "golang.org/x/image/webp" // register WebP decoder
)

const maxDimension = 1920

// ProcessResult holds the re-encoded image bytes and its MIME type.
type ProcessResult struct {
	Data        []byte
	ContentType string
}

// Process reads an image from r, detects whether it is JPEG, PNG, or WebP,
// resizes it so neither dimension exceeds maxDimension, and re-encodes it.
// JPEG and PNG are preserved in their original format; WebP is converted to PNG.
// It returns an error for unsupported types.
func Process(r io.Reader) (*ProcessResult, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("image.Process: read: %w", err)
	}

	ct := http.DetectContentType(raw)
	var outCT string
	switch ct {
	case "image/jpeg":
		outCT = "image/jpeg"
	case "image/png":
		outCT = "image/png"
	case "image/webp":
		// WebP is decoded and re-encoded as PNG (the registered decoder handles it).
		outCT = "image/png"
	default:
		return nil, fmt.Errorf("image.Process: unsupported content type %q (only JPEG, PNG, and WebP accepted)", ct)
	}

	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("image.Process: decode: %w", err)
	}

	img = resize(img, maxDimension)

	var buf bytes.Buffer
	switch outCT {
	case "image/jpeg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, fmt.Errorf("image.Process: jpeg encode: %w", err)
		}
	case "image/png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("image.Process: png encode: %w", err)
		}
	}

	return &ProcessResult{Data: buf.Bytes(), ContentType: outCT}, nil
}

// resize returns img scaled down so neither side exceeds max, preserving
// aspect ratio. Returns img unchanged if it already fits within max.
func resize(img image.Image, maxSide int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxSide && h <= maxSide {
		return img
	}

	var newW, newH int
	if w > h {
		newW = maxSide
		newH = h * maxSide / w
	} else {
		newH = maxSide
		newW = w * maxSide / h
	}
	if newH < 1 {
		newH = 1
	}
	if newW < 1 {
		newW = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := range newH {
		for x := range newW {
			srcX := x * w / newW
			srcY := y * h / newH
			dst.Set(x, y, img.At(b.Min.X+srcX, b.Min.Y+srcY))
		}
	}
	return dst
}
