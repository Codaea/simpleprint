package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/makeworld-the-better-one/dither/v2"
	"github.com/mect/go-escpos"
)

func main() {

	_ = godotenv.Overload(".env")

	// printer setup
	printerPath, found := os.LookupEnv("PRINTER_PATH")
	if !found {
		printerPath = ""
	}
	p, err := escpos.NewUSBPrinterByPath(printerPath)
	if err != nil {
		fmt.Println("No Printa Found!!")
		fmt.Println("Failed to connect to printer:", err)
		return
	}
	p.Init()
	p.Smooth(true)

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(func(c *gin.Context) {
		c.Set("printer", p)
		c.Next()
	})

	router.POST("/print", handlePrint)

	fmt.Printf("Listening and serving on 0.0.0.0:%s\n", os.Getenv("PORT"))
	router.Run() // listen and serve on 0.0.0.0:8080 (or whatever is set as PORT environment variable)
}

func processImage(image Image) image.Image {

	palette := []color.Color{
		color.Black, color.White,
	}

	d := dither.NewDitherer(palette)

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(image.Data))
	img, err := png.Decode(reader)
	if err != nil {
		fmt.Printf("failed to decode image: %v", err)
		return nil
	}

	// Create a new image from the decoded data

	// Apply dither mode if specified
	switch image.DitherMode {
	case FloydSteinberg:
		d.Matrix = dither.FloydSteinberg
	case None:
		return img // No dithering, return original image
	}

	return d.Dither(img)
}

/*
Let's define a JSON schema for what we need

{
	"receipt": [
		{
			"type": "line",
			"content": "Hello World",
			"font-size": 2,
			"font": "A",
			"alignment": "center",
			"underline": false
		},
		{
			"type": "feed",
			"lines": 2
		},
		{
			"type": "barcode",
			"code": "123456789012",
			"barcode-type": "CODE128"
		},
		{
			"type": "qr",
			"code": "https://example.com",
			"size": 8
		},
		{
			"type": "image",
			"data": "data:image/png;base64,iVBORw0KGgoAAAANSU...",
			"alignment": "center",
			"dither-mode": "floydsteinberg"
		},
		{
			"type": "text",
			"content": "Multi-line text\nwith newlines",
			"font-size": 1,
			"font": "A",
			"alignment": "left",
			"underline": false
		}
	]
}
*/
