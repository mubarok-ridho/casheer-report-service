package utils

import (
	"fmt"
	"time"

	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile"
)

type BluetoothPrinter struct {
	Device *api.Device
}

// Connect to bluetooth printer
func ConnectToPrinter(macAddress string) (*BluetoothPrinter, error) {
	// Discover and connect to device
	adapter, err := api.GetDefaultAdapter()
	if err != nil {
		return nil, fmt.Errorf("failed to get bluetooth adapter: %v", err)
	}

	devices, err := adapter.GetDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %v", err)
	}

	var targetDevice *api.Device
	for _, device := range devices {
		props, err := device.GetProperties()
		if err != nil {
			continue
		}
		if props.Address == macAddress {
			targetDevice = device
			break
		}
	}

	if targetDevice == nil {
		return nil, fmt.Errorf("printer with MAC %s not found", macAddress)
	}

	// Connect to device
	err = targetDevice.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	// Wait for connection
	time.Sleep(2 * time.Second)

	return &BluetoothPrinter{
		Device: targetDevice,
	}, nil
}

// Print raw bytes
func PrintToBluetooth(macAddress string, data []byte) error {
	printer, err := ConnectToPrinter(macAddress)
	if err != nil {
		return err
	}
	defer printer.Device.Disconnect()

	// Find the Serial Port Profile (SPP) service
	services, err := printer.Device.GetServices()
	if err != nil {
		return fmt.Errorf("failed to get services: %v", err)
	}

	for _, servicePath := range services {
		service, err := profile.NewGattService1(servicePath)
		if err != nil {
			continue
		}

		// Look for a characteristic that can write
		chars, err := service.GetCharacteristics()
		if err != nil {
			continue
		}

		for _, charPath := range chars {
			char, err := profile.NewGattCharacteristic1(charPath)
			if err != nil {
				continue
			}

			props, err := char.GetProperties()
			if err != nil {
				continue
			}

			// Check if characteristic supports write
			if props.Flags&profile.FlagWrite != 0 {
				// Write data to printer
				err = char.WriteValue(data, nil)
				if err == nil {
					return nil
				}
			}
		}
	}

	return fmt.Errorf("no suitable write characteristic found")
}

// ESC/POS Commands
var ESC = byte(0x1B)
var GS = byte(0x1D)

// Initialize printer
func InitPrinter() []byte {
	return []byte{ESC, 0x40} // ESC @
}

// Print and feed lines
func PrintAndFeed(lines byte) []byte {
	return []byte{ESC, 0x64, lines} // ESC d n
}

// Select cut mode and cut paper
func CutPaper() []byte {
	return []byte{GS, 0x56, 0x41, 0x00} // GS V m
}

// Set alignment
func SetAlignment(align byte) []byte {
	// align: 0 = left, 1 = center, 2 = right
	return []byte{ESC, 0x61, align} // ESC a n
}

// Set bold
func SetBold(bold bool) []byte {
	if bold {
		return []byte{ESC, 0x45, 0x01} // ESC E 1
	}
	return []byte{ESC, 0x45, 0x00} // ESC E 0
}

// Set font size
func SetFontSize(width, height byte) []byte {
	// width/height: 0 = normal, 1 = 2x, 2 = 3x, etc
	return []byte{GS, 0x21, (height << 4) | width} // GS ! n
}

// Print QR Code
func PrintQRCode(data string) []byte {
	var cmd []byte

	// Set QR code model
	cmd = append(cmd, GS, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x50, 0x30)

	// Set QR code size
	cmd = append(cmd, GS, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, 0x08)

	// Store QR code data
	nLow := byte(len(data) & 0xFF)
	nHigh := byte((len(data) >> 8) & 0xFF)
	cmd = append(cmd, GS, 0x28, 0x6B, nLow, nHigh, 0x31, 0x50, 0x30)
	cmd = append(cmd, []byte(data)...)

	// Print QR code
	cmd = append(cmd, GS, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30)

	return cmd
}
