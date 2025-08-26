package main

import (
	"encoding/json"
	"fmt"

	"github.com/mect/go-escpos"
)

type PrintRequest struct {
	Receipt []ReceiptItem `json:"receipt"`
}

// ReceiptItem represents any type of item that can appear on a receipt
type ReceiptItem interface{}

// RawReceiptItem is used for JSON unmarshaling with type discrimination
type RawReceiptItem struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:",inline"`
}

// Custom unmarshaling for PrintRequest
func (pr *PrintRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		Receipt []json.RawMessage `json:"receipt"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	pr.Receipt = make([]ReceiptItem, len(raw.Receipt))

	for i, itemData := range raw.Receipt {
		// First, extract the type field
		var typeExtractor struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(itemData, &typeExtractor); err != nil {
			return fmt.Errorf("error extracting type from item %d: %v", i, err)
		}

		// Then unmarshal the full object based on type
		switch typeExtractor.Type {
		case "line":
			var line Line
			if err := json.Unmarshal(itemData, &line); err != nil {
				return fmt.Errorf("error unmarshaling line: %v", err)
			}
			pr.Receipt[i] = line
		case "barcode":
			var barcode Barcode
			if err := json.Unmarshal(itemData, &barcode); err != nil {
				return fmt.Errorf("error unmarshaling barcode: %v", err)
			}
			pr.Receipt[i] = barcode
		case "qr":
			var qr QRCode
			if err := json.Unmarshal(itemData, &qr); err != nil {
				return fmt.Errorf("error unmarshaling qr: %v", err)
			}
			pr.Receipt[i] = qr
		case "image":
			var image Image
			if err := json.Unmarshal(itemData, &image); err != nil {
				return fmt.Errorf("error unmarshaling image: %v", err)
			}
			pr.Receipt[i] = image
		case "text":
			var text Text
			if err := json.Unmarshal(itemData, &text); err != nil {
				return fmt.Errorf("error unmarshaling text: %v", err)
			}
			pr.Receipt[i] = text
		case "feed":
			var feed Feed
			if err := json.Unmarshal(itemData, &feed); err != nil {
				return fmt.Errorf("error unmarshaling feed: %v", err)
			}
			pr.Receipt[i] = feed
		default:
			return fmt.Errorf("unknown receipt item type: %s", typeExtractor.Type)
		}
	}

	return nil
}

type FontType string

const (
	FontA FontType = "A"
	FontB FontType = "B"
	FontC FontType = "C"
)

type AlignmentType string

const (
	AlignLeft   AlignmentType = "left"
	AlignRight  AlignmentType = "right"
	AlignCenter AlignmentType = "center"
)

func (a *AlignmentType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "left", "right", "center":
		*a = AlignmentType(s)
		return nil
	default:
		return fmt.Errorf("invalid alignment: %s. Must be left, right, or center", s)
	}
}

func (a AlignmentType) ToEscposAlignment() escpos.Alignment {
	switch a {
	case AlignLeft:
		return escpos.AlignLeft
	case AlignRight:
		return escpos.AlignRight
	case AlignCenter:
		return escpos.AlignCenter
	default:
		return escpos.AlignLeft // default
	}
}

type BarcodeType string

const (
	BarcodeUPCA    BarcodeType = "UPCA"
	BarcodeUPCE    BarcodeType = "UPCE"
	BarcodeEAN13   BarcodeType = "EAN13"
	BarcodeEAN8    BarcodeType = "EAN8"
	BarcodeCODE39  BarcodeType = "CODE39"
	BarcodeCODE128 BarcodeType = "CODE128"
)

func (b *BarcodeType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "UPCA", "UPCE", "EAN13", "EAN8", "CODE39", "CODE128":
		*b = BarcodeType(s)
		return nil
	default:
		return fmt.Errorf("invalid barcode type: %s", s)
	}
}

// this gets called by the handler when converting to ESC/POS so it uses that packages types
func (t BarcodeType) ToEscposBarcodeType() escpos.BarcodeType {
	switch t {
	case "CODABAR":
		return escpos.BarcodeTypeCODABAR
	case "CODE128":
		return escpos.BarcodeTypeCODE128
	case "CODE39":
		return escpos.BarcodeTypeCODE39
	case "EAN13":
		return escpos.BarcodeTypeEAN13
	case "EAN8":
		return escpos.BarcodeTypeEAN8
	case "ITF":
		return escpos.BarcodeTypeITF
	case "UPCA":
		return escpos.BarcodeTypeUPCA
	case "UPCE":
		return escpos.BarcodeTypeUPCE
	}
	return escpos.BarcodeTypeCODABAR // shits fucked if we are here

}

type DitherMode string

const (
	None           DitherMode = "none"
	FloydSteinberg DitherMode = "floydsteinberg"
)

func (d *DitherMode) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "none", "floydsteinberg":
		*d = DitherMode(s)
		return nil
	default:
		return fmt.Errorf("invalid dither mode: %s", s)
	}
}

// Helper functions to convert custom types to escpos types
func (f FontType) ToEscposFont() escpos.Font {
	switch f {
	case FontA:
		return escpos.FontA
	case FontB:
		return escpos.FontB
	case FontC:
		return escpos.FontC
	default:
		return escpos.FontA // default
	}
}
