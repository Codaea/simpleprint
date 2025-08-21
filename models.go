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
			pr.Receipt[i] = feed // Just assign the feed directly
		default:
			return fmt.Errorf("unknown receipt item type: %s", typeExtractor.Type)
		}
	}

	return nil
}

// All receipt item types implement ReceiptItem
var (
	_ ReceiptItem = (*Line)(nil)
	_ ReceiptItem = (*Barcode)(nil)
	_ ReceiptItem = (*QRCode)(nil)
	_ ReceiptItem = (*Image)(nil)
	_ ReceiptItem = (*Text)(nil)
	_ ReceiptItem = (*Feed)(nil)
)

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

func (t BarcodeType) ToEscposBarcodeType() escpos.BarcodeType {
	switch t {
	case "codabar":
		return escpos.BarcodeTypeCODABAR
	case "code128":
		return escpos.BarcodeTypeCODE128
	case "code39":
		return escpos.BarcodeTypeCODE39
	case "ean13":
		return escpos.BarcodeTypeEAN13
	case "ean8":
		return escpos.BarcodeTypeEAN8
	case "itf":
		return escpos.BarcodeTypeITF
	case "upca":
		return escpos.BarcodeTypeUPCA
	case "upce":
		return escpos.BarcodeTypeUPCE
	}

	return escpos.BarcodeTypeCODABAR // shits fucked if we are here

}

type Line struct {
	Type      string        `json:"type"`
	Content   string        `json:"content"`
	FontSize  int           `json:"font-size"`
	Font      FontType      `json:"font"`
	Alignment AlignmentType `json:"alignment"`
	Underline bool          `json:"underline"`
}

// Custom unmarshaling for Line to handle string/number flexibility
func (l *Line) UnmarshalJSON(data []byte) error {
	type Alias Line
	aux := &struct {
		FontSize  interface{} `json:"font-size"`
		Underline interface{} `json:"underline"`
		*Alias
	}{
		Alias: (*Alias)(l),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle FontSize (can be string or number)
	switch v := aux.FontSize.(type) {
	case string:
		if parsed, err := json.Number(v).Int64(); err == nil {
			l.FontSize = int(parsed)
		} else {
			return fmt.Errorf("invalid font-size: %s", v)
		}
	case float64:
		l.FontSize = int(v)
	case int:
		l.FontSize = v
	}

	// Handle Underline (can be string or boolean)
	switch v := aux.Underline.(type) {
	case string:
		l.Underline = v == "true"
	case bool:
		l.Underline = v
	}

	return nil
}

type Barcode struct {
	Type        string      `json:"type"`
	Code        string      `json:"code"`
	BarcodeType BarcodeType `json:"barcode_type"`
}

type QRCode struct {
	Type string `json:"type"`
	Code string `json:"code"`
	Size int    `json:"size"`
}

type Image struct {
	Type       string        `json:"type"`
	Data       string        `json:"data"`
	Alignment  AlignmentType `json:"alignment"`
	DitherMode DitherMode    `json:"dither_mode"`
}

type Text struct {
	Type      string        `json:"type"`
	Content   string        `json:"content"`
	FontSize  int           `json:"font_size"`
	Font      FontType      `json:"font"`
	Alignment AlignmentType `json:"alignment"`
	Underline bool          `json:"underline"`
}

// Custom unmarshaling for Text to handle string/number flexibility
func (t *Text) UnmarshalJSON(data []byte) error {
	type Alias Text
	aux := &struct {
		FontSize  interface{} `json:"font_size"`
		Underline interface{} `json:"underline"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle FontSize (can be string or number)
	switch v := aux.FontSize.(type) {
	case string:
		if parsed, err := json.Number(v).Int64(); err == nil {
			t.FontSize = int(parsed)
		} else {
			return fmt.Errorf("invalid font_size: %s", v)
		}
	case float64:
		t.FontSize = int(v)
	case int:
		t.FontSize = v
	}

	// Handle Underline (can be string or boolean)
	switch v := aux.Underline.(type) {
	case string:
		t.Underline = v == "true"
	case bool:
		t.Underline = v
	}

	return nil
}

type Feed struct {
	Type  string `json:"type"`
	Lines int    `json:"lines"`
}
