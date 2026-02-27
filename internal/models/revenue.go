package models

import (
	"time"

	"gorm.io/gorm"
)

type Revenue struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	TenantID     uint           `json:"tenant_id" gorm:"not null;uniqueIndex:idx_tenant_date"`
	Date         time.Time      `json:"date" gorm:"not null;uniqueIndex:idx_tenant_date"`
	TotalRevenue float64        `json:"total_revenue" gorm:"default:0"`
	TotalExpense float64        `json:"total_expense" gorm:"default:0"`
	NetRevenue   float64        `json:"net_revenue" gorm:"default:0"`
	OrderCount   int            `json:"order_count" gorm:"default:0"`
	ExpenseCount int            `json:"expense_count" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type RevenueEvent struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	TenantID    uint      `json:"tenant_id" gorm:"not null;index"`
	Type        string    `json:"type"`         // "order" atau "expense"
	ReferenceID uint      `json:"reference_id"` // ID dari order atau expense
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
}
