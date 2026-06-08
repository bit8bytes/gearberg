// Package barcodes generates QR code and barcode images for serialized units.
package barcodes

import (
	"bytes"
	"fmt"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
)

const (
	qrSize   = 256
	barcodeW = 400
	barcodeH = 100
)

// QR returns a PNG-encoded QR code image for the given id string.
func QR(id string) ([]byte, error) {
	bc, err := qr.Encode(id, qr.M, qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("barcodes.QR: encode: %w", err)
	}
	bc, err = barcode.Scale(bc, qrSize, qrSize)
	if err != nil {
		return nil, fmt.Errorf("barcodes.QR: scale: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, bc); err != nil {
		return nil, fmt.Errorf("barcodes.QR: png: %w", err)
	}
	return buf.Bytes(), nil
}

// Code128 returns a PNG-encoded Code128 barcode image for the given id string.
func Code128(id string) ([]byte, error) {
	raw, err := code128.Encode(id)
	if err != nil {
		return nil, fmt.Errorf("barcodes.Code128: encode: %w", err)
	}
	scaled, err := barcode.Scale(raw, barcodeW, barcodeH)
	if err != nil {
		return nil, fmt.Errorf("barcodes.Code128: scale: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, fmt.Errorf("barcodes.Code128: png: %w", err)
	}
	return buf.Bytes(), nil
}
