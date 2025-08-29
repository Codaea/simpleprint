package main

import (
	"fmt"
	"image"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/mect/go-escpos"
)

// Types for receipt items - defined close to where they're used
type Line struct {
	Type      string        `json:"type"`
	Content   string        `json:"content"`
	FontSize  int           `json:"font_size"`
	Font      FontType      `json:"font"`
	Alignment AlignmentType `json:"alignment"`
	Underline bool          `json:"underline"`
}

type Text struct {
	Type      string        `json:"type"`
	Content   string        `json:"content"`
	FontSize  int           `json:"font_size"`
	Font      FontType      `json:"font"`
	Alignment AlignmentType `json:"alignment"`
	Underline bool          `json:"underline"`
}

type Feed struct {
	Type  string `json:"type"`
	Lines int    `json:"lines"`
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
	DitherMode string        `json:"dither_mode"`
	img        image.Image   // decoded image, not directly unmarshaled
}

// Global mutex to serialize printer access
var printerMutex sync.Mutex

func handlePrint(c *gin.Context) {
	// Try to lock the printer, return busy if already in use
	if !printerMutex.TryLock() {
		c.JSON(503, gin.H{
			"error":   "Printer is busy",
			"message": "Another print job is currently in progress. Please try again later.",
		})
		return
	}
	defer printerMutex.Unlock()

	p := c.MustGet("printer").(*escpos.Printer) // get printer object from middleware

	var req PrintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(req.Receipt)

	// Process each receipt item
	for _, item := range req.Receipt {
		fmt.Printf("Printing Line %v\n", item)
		switch v := item.(type) {
		case Line:
			// Print line
			p.Font(v.Font.ToEscposFont())
			p.Align(v.Alignment.ToEscposAlignment())
			p.Size(uint8(v.FontSize), uint8(v.FontSize))
			p.Underline(v.Underline)
			p.PrintLn(v.Content)
		case Text:
			// Print text (similar to line)
			p.Font(v.Font.ToEscposFont())
			p.Align(v.Alignment.ToEscposAlignment())
			p.Size(uint8(v.FontSize), uint8(v.FontSize))
			p.Underline(v.Underline)
			p.Print(v.Content)
		case Feed:
			// Feed lines
			p.Feed(v.Lines)
		case Barcode:
			p.Align(escpos.AlignCenter)
			err := p.Barcode(v.Code, v.BarcodeType.ToEscposBarcodeType())
			if err != nil {
				fmt.Printf("Error printing barcode: %v\n", err)
			}
		case QRCode:
			// Print QR code
			p.Align(escpos.AlignCenter)
			p.QR(v.Code, v.Size)
		case Image:
			// Print image (you'll need to decode base64 and process)
			p.Align(v.Alignment.ToEscposAlignment())
			p.Image(processImage(v))
		}
	}

	p.Cut()
	c.JSON(200, gin.H{"success": true})
}
