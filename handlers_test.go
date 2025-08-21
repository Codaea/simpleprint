package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mect/go-escpos"
	"github.com/stretchr/testify/assert"
)

// Mock printer for testing
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	printer, err := escpos.NewUSBPrinterByPath("")
	if err != nil {
		panic("Failed to create mock printer: " + err.Error())
	}
	printer.Init()
	printer.Smooth(true)

	// Mock printer middleware
	r.Use(func(c *gin.Context) {
		// You'll need to create a real printer connection here for e2e tests

		c.Set("printer", printer)
		c.Next()
	})

	r.POST("/print", handlePrint)
	return r
}

func TestHandlePrint_BasicLine(t *testing.T) {
	router := setupTestRouter()

	body := `
	{
		"receipt": [
		{
		"type": "line",
		"content": "Hello, World!",
		"font": "A",
		"alignment": "center",
		"font_size": 1,
		"underline": false
		},
		{
			"type": "feed",
			"lines": 2
		}
		]
	}
	`
	req, _ := http.NewRequest("POST", "/print", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["success"].(bool))
}

func TestHandlePrint_MultipleItems(t *testing.T) {
	router := setupTestRouter()

	body := `
	{
		"receipt": [
			{
				"type": "line",
				"content": "RECEIPT HEADER",
				"font": "A",
				"alignment": "center",
				"font_size": 2,
				"underline": true
			},
			{
				"type": "feed",
				"lines": 1
			},
			{
				"type": "line",
				"content": "Item 1: $10.00",
				"font": "B",
				"alignment": "left",
				"font_size": 1,
				"underline": false
			},
			{
				"type": "line",
				"content": "Item 2: $15.00",
				"font": "B",
				"alignment": "left",
				"font_size": 1,
				"underline": false
			},
			{
				"type": "feed",
				"lines": 1
			},
			{
				"type": "line",
				"content": "Total: $25.00",
				"font": "A",
				"alignment": "right",
				"font_size": 1,
				"underline": true
			},
			{
				"type": "feed",
				"lines": 2
			}
		]
	}
	`
	req, _ := http.NewRequest("POST", "/print", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestHandlePrint_WithBarcode(t *testing.T) {
	router := setupTestRouter()

	body := `
	{
		"receipt": [
			{
				"type": "line",
				"content": "Product Info",
				"font": "A",
				"alignment": "center",
				"font_size": 1,
				"underline": false
			},
			{
				"type": "feed",
				"lines": 1
			},
			{
				"type": "barcode",
				"code": "123456789012",
				"barcode_type": "CODE128"
			},
			{
				"type": "feed",
				"lines": 2
			}
		]
	}
	`

	req, _ := http.NewRequest("POST", "/print", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestHandlePrint_WithQRCode(t *testing.T) {
	router := setupTestRouter()

	body := `
	{
	"receipt": [
		{
			"type": "qr",
			"code": "https://example.com/receipt/12345",
			"size": 16
		}
	]
}
	`
	req, _ := http.NewRequest("POST", "/print", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestHandlePrint_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("POST", "/print", bytes.NewBufferString(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "error")
}

func TestHandlePrint_PrinterBusy(t *testing.T) {
	router := setupTestRouter()

	// Lock the printer mutex to simulate busy printer
	printerMutex.Lock()

	receipt := PrintRequest{
		Receipt: []ReceiptItem{
			Line{Content: "Test", Font: FontA, Alignment: AlignLeft, FontSize: 1},
		},
	}

	body, _ := json.Marshal(receipt)
	req, _ := http.NewRequest("POST", "/print", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 503, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Printer is busy", response["error"])

	// Unlock for cleanup
	printerMutex.Unlock()
}

func TestHandlePrint_ConcurrentRequests(t *testing.T) {
	router := setupTestRouter()

	receipt := PrintRequest{
		Receipt: []ReceiptItem{
			Line{Content: "Concurrent Test", Font: FontA, Alignment: AlignCenter, FontSize: 1},
			Feed{Lines: 3},
		},
	}

	var wg sync.WaitGroup
	results := make(chan int, 5)

	// Send 5 concurrent requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			body, _ := json.Marshal(receipt)
			req, _ := http.NewRequest("POST", "/print", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			results <- w.Code
		}(i)
	}

	wg.Wait()
	close(results)

	successCount := 0
	busyCount := 0

	for code := range results {
		switch code {
		case 200:
			successCount++
		case 503:
			busyCount++
		}
	}

	// At least one should succeed, others should be busy
	assert.True(t, successCount >= 1)
	assert.True(t, busyCount >= 1)
	assert.Equal(t, 5, successCount+busyCount)
}

// Helper function for complete receipt test
func TestHandlePrint_CompleteReceipt(t *testing.T) {
	router := setupTestRouter()

	body := `
	{
		"receipt": [
			{
				"type": "line",
				"content": "STORE NAME",
				"font": "A",
				"alignment": "center",
				"font-size": 2,
				"underline": true
			},
			{
				"type": "line",
				"content": "123 Main St, City, State",
				"font": "B",
				"alignment": "center",
				"font-size": 1,
				"underline": false
			},
			{
				"type": "line",
				"content": "Tel: (555) 123-4567",
				"font": "B",
				"alignment": "center",
				"font-size": 1,
				"underline": false
			},
			{
				"type": "feed",
				"lines": 2
			},
			{
				"type": "line",
				"content": "RECEIPT #12345",
				"font": "A",
				"alignment": "center",
				"font-size": 1,
				"underline": false
			},
			{
				"type": "feed",
				"lines": 1
			},
			{
				"type": "text",
				"content": "Item 1",
				"font": "B",
				"alignment": "left",
				"font-size": 1,
				"underline": false
			},
			{
				"type": "text",
				"content": "$10.00",
				"font": "B",
				"alignment": "right",
				"font-size": 1,
				"underline": false
			},
			{
				"type": "text",
				"content": "Item 2",
				"font": "B",
				"alignment": "left",
				"font-size": 1,
				"underline": false
			},
			{
				"type": "text",
				"content": "$15.00",
				"font": "B",
				"alignment": "right",
				"font-size": 1,
				"underline": false
			},
			{
				"type": "feed",
				"lines": 1
			},
			{
				"type": "line",
				"content": "--------------------------------",
				"font": "B",
				"alignment": "center",
				"font-size": 1,
				"underline": false
			},
			{
				"type": "line",
				"content": "TOTAL: $25.00",
				"font": "A",
				"alignment": "center",
				"font-size": 1,
				"underline": true
			},
			{
				"type": "feed",
				"lines": 2
			},
			{
				"type": "qr",
				"code": "https://store.com/receipt/12345",
				"size": 4
			},
			{
				"type": "feed",
				"lines": 3
			}
		]
	}
	`

	req, _ := http.NewRequest("POST", "/print", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Add a small delay to allow printer to finish
	time.Sleep(100 * time.Millisecond)
}
