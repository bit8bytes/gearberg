// Copyright (C) 2026 Tobias Gleiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
