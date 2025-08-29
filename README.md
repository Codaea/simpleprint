# SimplePrint

A lightweight HTTP API server for thermal receipt printers using ESC/POS protocol. SimplePrint provides a simple REST interface to control thermal printers for receipt printing, QR codes, barcodes, and images.

## Features

- 🖨️ **Thermal Printer Support** - Works with ESC/POS compatible thermal printers via USB
- 📄 **Multiple Content Types** - Print text, QR codes, barcodes, and images
- 🔧 **Flexible Formatting** - Control fonts, sizes, alignment, and styling
- 🔒 **Thread-Safe** - Handles concurrent print requests safely
- 🚀 **Easy Integration** - Simple HTTP REST API

## Installation

### Prerequisites

- Go 1.23.5 or later
- ESC/POS compatible thermal printer connected via USB

### Build from Source

```bash
git clone https://github.com/codaea/simpleprint.git
cd simpleprint
go mod download
go build -o simpleprint

```

### Configuration
SimplePrint uses environment variables for configuration. You can set these in a .env file in the project root.

| Variable      | Default   | Description                                      |
|---------------|-----------|--------------------------------------------------|
| `GIN_MODE`    | release   | Gin server mode (`release` or `debug`)           |
| `PORT`        | 3000      | Port for the HTTP server                         |
| `PRINTER_PATH`| (empty)   | USB path to the printer (optional, Linux only)   |

Copy `.env.sample` to `.env` and adjust as needed.


### Run the Server

```bash
./simpleprint
```

The server will start on port `5010` by default.

## API Reference

### Base URL
```
http://localhost:5010
```

### Print Receipt

**Endpoint:** `POST /print`

**Description:** Sends a print job to the thermal printer. The request contains an array of print commands that will be executed sequentially.

**Request Body:**
```json
{
  "receipt": [
    {
      "type": "line",
      "content": "Restaurant Name",
      "font_size": 4,
      "font": "A",
      "alignment": "center",
      "underline": false
    },
    {
      "type": "qr",
      "code": "https://example.com/receipt/12345",
      "size": 8
    }
  ]
}
```

The `receipt` field is an array of print command objects. Each command represents a different element to print.

## Print Command Types

### Text Line (`line`)

Prints a single line of text with formatting options.

```json
{
  "type": "line",
  "content": "Hello World",
  "font_size": 2,
  "font": "A",
  "alignment": "center",
  "underline": false
}
```

**Parameters:**
- `content` (string): The text to print
- `font_size` (integer): Font size multiplier (1-8)
- `font` (string): Font type - `"A"`, `"B"`, or `"C"`
- `alignment` (string): Text alignment - `"left"`, `"center"`, or `"right"`
- `underline` (boolean): Whether to underline the text

### Multi-line Text (`text`)

Prints text that can span multiple lines (supports `\n` newlines).

```json
{
  "type": "text",
  "content": "Multi-line text\nwith newlines\nSupported here",
  "font_size": 1,
  "font": "A",
  "alignment": "left",
  "underline": false
}
```

**Parameters:** Same as `line` type, but `content` can contain newline characters.

### Line Feed (`feed`)

Advances the paper by the specified number of lines.

```json
{
  "type": "feed",
  "lines": 3
}
```

**Parameters:**
- `lines` (integer): Number of lines to feed

### QR Code (`qr`)

Prints a QR code with the specified data.

```json
{
  "type": "qr",
  "code": "https://example.com/order/12345",
  "size": 8
}
```

**Parameters:**
- `code` (string): The data to encode in the QR code
- `size` (integer): QR code size (1-16, where larger numbers create bigger codes)

### Barcode (`barcode`)

Prints a barcode with the specified data and format.

```json
{
  "type": "barcode",
  "code": "123456789012",
  "barcode_type": "CODE128"
}
```

**Parameters:**
- `code` (string): The data to encode in the barcode
- `barcode_type` (string): Barcode format - `"UPCA"`, `"UPCE"`, `"EAN13"`, `"EAN8"`, `"CODE39"`, or `"CODE128"`

### Image (`image`)

Prints an image from base64-encoded data.

```json
{
  "type": "image",
  "data": "data:image/png;base64,iVBORw0KGgoAAAANSU...",
  "alignment": "center",
  "dither_mode": "floydsteinberg"
}
```

**Parameters:**
- `data` (string): Base64-encoded image data (PNG format recommended)
- `alignment` (string): Image alignment - `"left"`, `"center"`, or `"right"`
- `dither-mode` (string): Dithering algorithm - `"none"` or `"floydsteinberg"`

## Response Codes

### Success Response
**Status Code:** `200 OK`
```json
{
  "success": true
}
```

### Error Responses

**Status Code:** `400 Bad Request`
```json
{
  "error": "Invalid JSON format or missing required fields"
}
```

**Status Code:** `503 Service Unavailable`
```json
{
  "error": "Printer is busy",
  "message": "Another print job is currently in progress. Please try again later."
}
```

## Complete Example

Here's a complete example that demonstrates printing a receipt with multiple elements:

```json
{
  "receipt": [
    {
      "type": "line",
      "content": "RECEIPT",
      "font_size": 4,
      "font": "A",
      "alignment": "center",
      "underline": true
    },
    {
      "type": "feed",
      "lines": 1
    },
    {
      "type": "line",
      "content": "Coffee Shop",
      "font_size": 2,
      "font": "A",
      "alignment": "center",
      "underline": false
    },
    {
      "type": "line",
      "content": "123 Main St, City",
      "font_size": 1,
      "font": "A",
      "alignment": "center",
      "underline": false
    },
    {
      "type": "feed",
      "lines": 2
    },
    {
      "type": "text",
      "content": "Order #12345\nDate: 2025-08-20\nTime: 14:30",
      "font_size": 1,
      "font": "A",
      "alignment": "left",
      "underline": false
    },
    {
      "type": "feed",
      "lines": 1
    },
    {
      "type": "line",
      "content": "--------------------------------",
      "font_size": 1,
      "font": "A",
      "alignment": "center",
      "underline": false
    },
    {
      "type": "text",
      "content": "1x Latte         $4.50\n1x Croissant     $3.25\nTax              $0.62",
      "font_size": 1,
      "font": "A",
      "alignment": "left",
      "underline": false
    },
    {
      "type": "line",
      "content": "--------------------------------",
      "font_size": 1,
      "font": "A",
      "alignment": "center",
      "underline": false
    },
    {
      "type": "line",
      "content": "TOTAL: $8.37",
      "font_size": 2,
      "font": "A",
      "alignment": "center",
      "underline": false
    },
    {
      "type": "feed",
      "lines": 2
    },
    {
      "type": "qr",
      "code": "https://coffeeshop.com/receipt/12345",
      "size": 6
    },
    {
      "type": "feed",
      "lines": 1
    },
    {
      "type": "line",
      "content": "Thank you for your visit!",
      "font_size": 1,
      "font": "A",
      "alignment": "center",
      "underline": false
    }
  ]
}
```

## Usage with cURL

```bash
curl -X POST http://localhost:5010/print \
  -H "Content-Type: application/json" \
  -d '{
    "receipt": [
      {
        "type": "line",
        "content": "Hello World!",
        "font_size": 2,
        "font": "A",
        "alignment": "center",
        "underline": false
      }
    ]
  }'
```

## Troubleshooting

### Printer Not Found
If you see "No Printa Found!!" when starting the server, ensure:
- Your thermal printer is connected via USB
- The printer is powered on
- You have the necessary permissions to access USB devices
- The printer uses ESC/POS protocol

### Print Jobs Hanging
If print jobs seem to hang:
- Check that the printer has paper
- Ensure the printer is not in an error state (paper jam, cover open, etc.)
- Restart the SimplePrint server

## Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [mect/go-escpos](https://github.com/mect/go-escpos) - ESC/POS printer library
- [makeworld-the-better-one/dither](https://github.com/makeworld-the-better-one/dither) - Image dithering

## License

This project is open source. Please check the license file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
