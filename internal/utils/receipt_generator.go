package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/mubarok-ridho/casheer-report-service/internal/models"
)

type ReceiptGenerator struct{}

// Generate receipt based on template
func GenerateReceipt(tenantInfo map[string]interface{}, orderData map[string]interface{}, template *models.ReceiptTemplate) ([]byte, error) {
	var receipt []byte

	// Initialize printer
	receipt = append(receipt, InitPrinter()...)

	// Set alignment center for header
	receipt = append(receipt, SetAlignment(1)...)

	// Add logo if enabled
	if template.ShowLogo {
		receipt = append(receipt, []byte("[LOGO]\n")...) // Placeholder, actual logo would be ESC/POS image commands
	}

	// Store name
	receipt = append(receipt, SetBold(true)...)
	receipt = append(receipt, []byte(fmt.Sprintf("%s\n", tenantInfo["name"]))...)
	receipt = append(receipt, SetBold(false)...)

	// Store address and phone
	receipt = append(receipt, SetAlignment(0)...)
	receipt = append(receipt, []byte(fmt.Sprintf("%s\n", tenantInfo["address"]))...)
	receipt = append(receipt, []byte(fmt.Sprintf("Telp: %s\n", tenantInfo["phone"]))...)

	// Receipt header from template
	if template.Header != "" {
		receipt = append(receipt, []byte(template.Header)...)
		receipt = append(receipt, '\n')
	}

	// Separator
	receipt = append(receipt, []byte(strings.Repeat("-", template.CharactersPerLine))...)
	receipt = append(receipt, '\n')

	// Order info
	receipt = append(receipt, []byte(fmt.Sprintf("No: %s\n", orderData["order_number"]))...)
	receipt = append(receipt, []byte(fmt.Sprintf("Tgl: %s\n", orderData["date"]))...)
	receipt = append(receipt, []byte(fmt.Sprintf("Kasir: %s\n", "Admin"))...) // Would get from actual data

	// Separator
	receipt = append(receipt, []byte(strings.Repeat("-", template.CharactersPerLine))...)
	receipt = append(receipt, '\n')

	// Items header
	receipt = append(receipt, []byte(fmt.Sprintf("%-20s %2s %8s\n", "Item", "Qty", "Total"))...)
	receipt = append(receipt, []byte(strings.Repeat("-", template.CharactersPerLine))...)
	receipt = append(receipt, '\n')

	// Items
	items := orderData["items"].([]map[string]interface{})
	for _, item := range items {
		name := item["name"].(string)
		if len(name) > 18 {
			name = name[:18]
		}
		qty := item["quantity"].(int)
		subtotal := item["subtotal"].(float64)

		receipt = append(receipt, []byte(fmt.Sprintf("%-20s %2d %8.0f\n", name, qty, subtotal))...)

		// Add variation if any
		if notes, ok := item["notes"]; ok && notes != "" {
			receipt = append(receipt, []byte(fmt.Sprintf("  - %s\n", notes))...)
		}
	}

	// Separator
	receipt = append(receipt, []byte(strings.Repeat("-", template.CharactersPerLine))...)
	receipt = append(receipt, '\n')

	// Totals
	receipt = append(receipt, []byte(fmt.Sprintf("Subtotal: %35.0f\n", orderData["subtotal"]))...)

	if template.ShowTax {
		receipt = append(receipt, []byte(fmt.Sprintf("Pajak: %37.0f\n", orderData["tax"]))...)
	}

	receipt = append(receipt, SetBold(true)...)
	receipt = append(receipt, []byte(fmt.Sprintf("TOTAL: %36.0f\n", orderData["total"]))...)
	receipt = append(receipt, SetBold(false)...)

	receipt = append(receipt, []byte(fmt.Sprintf("Bayar: %35s\n", orderData["payment"]))...)

	// Separator
	receipt = append(receipt, []byte(strings.Repeat("=", template.CharactersPerLine))...)
	receipt = append(receipt, '\n')

	// Footer from template
	if template.Footer != "" {
		receipt = append(receipt, SetAlignment(1)...)
		receipt = append(receipt, []byte(template.Footer)...)
		receipt = append(receipt, '\n')
	}

	// Thank you message
	receipt = append(receipt, SetAlignment(1)...)
	receipt = append(receipt, []byte("Terima Kasih\n")...)
	receipt = append(receipt, []byte("Selamat Datang Kembali\n")...)

	// Date time
	receipt = append(receipt, SetAlignment(0)...)
	receipt = append(receipt, []byte(time.Now().Format("02/01/2006 15:04:05"))...)
	receipt = append(receipt, '\n')

	// Cut paper
	receipt = append(receipt, PrintAndFeed(3)...)
	receipt = append(receipt, CutPaper()...)

	return receipt, nil
}

// Generate test page
func GenerateTestPage(data map[string]interface{}) []byte {
	var test []byte

	test = append(test, InitPrinter()...)
	test = append(test, SetAlignment(1)...)
	test = append(test, SetBold(true)...)
	test = append(test, []byte(data["title"].(string))...)
	test = append(test, '\n', '\n')
	test = append(test, SetBold(false)...)

	test = append(test, SetAlignment(0)...)
	test = append(test, []byte(fmt.Sprintf("Tanggal: %s\n", data["date"]))...)
	test = append(test, []byte(fmt.Sprintf("Lebar Kertas: %s\n", data["width"]))...)
	test = append(test, '\n')
	test = append(test, []byte(data["content"].(string))...)
	test = append(test, '\n', '\n')

	// Print all characters
	test = append(test, []byte("!\"#$%&'()*+,-./0123456789:;<=>?@")...)
	test = append(test, '\n')
	test = append(test, []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")...)
	test = append(test, '\n')
	test = append(test, []byte("abcdefghijklmnopqrstuvwxyz")...)
	test = append(test, '\n', '\n')

	test = append(test, SetAlignment(1)...)
	test = append(test, []byte("PRINTER TEST OK")...)
	test = append(test, '\n', '\n')

	test = append(test, CutPaper()...)

	return test
}

// Generate QR Code receipt
func GenerateQRReceipt(tenantInfo map[string]interface{}, orderData map[string]interface{}, qrData string) []byte {
	var receipt []byte

	receipt = append(receipt, InitPrinter()...)
	receipt = append(receipt, SetAlignment(1)...)
	receipt = append(receipt, SetBold(true)...)
	receipt = append(receipt, []byte(fmt.Sprintf("%s\n", tenantInfo["name"]))...)
	receipt = append(receipt, SetBold(false)...)
	receipt = append(receipt, []byte(fmt.Sprintf("No: %s\n", orderData["order_number"]))...)
	receipt = append(receipt, []byte(fmt.Sprintf("Total: Rp%.0f\n", orderData["total"]))...)
	receipt = append(receipt, '\n')

	// Add QR Code
	receipt = append(receipt, PrintQRCode(qrData)...)
	receipt = append(receipt, '\n')

	receipt = append(receipt, []byte("Scan untuk pembayaran")...)
	receipt = append(receipt, '\n', '\n')

	receipt = append(receipt, CutPaper()...)

	return receipt
}
