package utils

import "fmt"

type BluetoothPrinter struct {
	MacAddress string
}

func ConnectToPrinter(macAddress string) (*BluetoothPrinter, error) {
	return nil, fmt.Errorf("bluetooth printing not supported on this platform")
}

func PrintToBluetooth(macAddress string, data []byte) error {
	return fmt.Errorf("bluetooth printing not supported on this platform")
}

// ESC/POS Commands
var ESC = byte(0x1B)
var GS = byte(0x1D)

func InitPrinter() []byte        { return []byte{ESC, 0x40} }
func PrintAndFeed(lines byte) []byte { return []byte{ESC, 0x64, lines} }
func CutPaper() []byte           { return []byte{GS, 0x56, 0x41, 0x00} }
func SetAlignment(align byte) []byte { return []byte{ESC, 0x61, align} }
func SetBold(bold bool) []byte {
	if bold { return []byte{ESC, 0x45, 0x01} }
	return []byte{ESC, 0x45, 0x00}
}
func SetFontSize(width, height byte) []byte { return []byte{GS, 0x21, (height << 4) | width} }
func PrintQRCode(data string) []byte {
	var cmd []byte
	cmd = append(cmd, GS, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x50, 0x30)
	cmd = append(cmd, GS, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, 0x08)
	nLow := byte(len(data) & 0xFF)
	nHigh := byte((len(data) >> 8) & 0xFF)
	cmd = append(cmd, GS, 0x28, 0x6B, nLow, nHigh, 0x31, 0x50, 0x30)
	cmd = append(cmd, []byte(data)...)
	cmd = append(cmd, GS, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30)
	return cmd
}
