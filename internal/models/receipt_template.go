package models

import (
	"time"

	"gorm.io/gorm"
)

type ReceiptTemplate struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	TenantID    *uint  `json:"tenant_id"` // Nullable untuk default templates
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`

	// Template Content
	Header       string `json:"header"` // HTML or plain text
	Footer       string `json:"footer"` // HTML or plain text
	ShowLogo     bool   `json:"show_logo" gorm:"default:true"`
	ShowTax      bool   `json:"show_tax" gorm:"default:false"`
	ShowDiscount bool   `json:"show_discount" gorm:"default:false"`

	// Format
	PaperWidth        string `json:"paper_width" gorm:"default:'58mm'"` // 58mm atau 80mm
	FontSize          int    `json:"font_size" gorm:"default:12"`
	CharactersPerLine int    `json:"characters_per_line" gorm:"default:32"`

	// Settings
	IsDefault bool           `json:"is_default" gorm:"default:false"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
