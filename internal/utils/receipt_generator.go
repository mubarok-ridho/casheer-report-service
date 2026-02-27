package utils

import (
	"fmt"
	"strings"
	"time"
)

func GenerateReceipt(tenant interface{}, template interface{}, orderID uint) []byte {
	// Format ESC/POS untuk printer thermal
	var receipt []byte

	// Init printer
	receipt = append(receipt, []byte{0x1B, 0x40}...) // ESC @ - Initialize printer

	// Header
	receipt = append(receipt, []byte(centerText(tenant.StoreName))...)
	receipt = append(receipt, '\n')
	receipt = append(receipt, []byte(strings.Repeat("-", 32))...)
	receipt = append(receipt, '\n')

	// Order info
	receipt = append(receipt, []byte(fmt.Sprintf("No: %d\n", orderID))...)
	receipt = append(receipt, []byte(fmt.Sprintf("Date: %s\n", time.Now().Format("02/01/2006 15:04"))...))
	receipt = append(receipt, []byte(strings.Repeat("-", 32))...)
	receipt = append(receipt, '\n')

	// Items
	// ... tambahkan items

	// Total
	receipt = append(receipt, []byte(strings.Repeat("-", 32))...)
	receipt = append(receipt, '\n')

	// Footer
	receipt = append(receipt, []byte(centerText("Terima Kasih"))...)
	receipt = append(receipt, []byte{0x1D, 0x56, 0x41, 0x03}...) // Cut paper

	return receipt
}

func centerText(text string) string {
	padding := (32 - len(text)) / 2
	if padding > 0 {
		return strings.Repeat(" ", padding) + text
	}
	return text
}
