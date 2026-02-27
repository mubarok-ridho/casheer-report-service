package models

import (
	"time"

	"gorm.io/gorm"
)

type Expense struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	TenantID      uint           `json:"tenant_id" gorm:"not null;index"`
	Category      string         `json:"category" gorm:"not null"` // Operational, Marketing, Salary, etc
	Description   string         `json:"description" gorm:"not null"`
	Amount        float64        `json:"amount" gorm:"not null"`
	Date          time.Time      `json:"date" gorm:"not null;index"`
	PaymentMethod string         `json:"payment_method"` // cash, transfer, etc
	Notes         string         `json:"notes"`
	ReceiptURL    string         `json:"receipt_url"` // foto bukti pengeluaran
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type ExpenseCategory struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	TenantID    uint      `json:"tenant_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Color       string    `json:"color"` // untuk UI
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
