package models

import (
	"time"
)

type DailyReport struct {
	ID       uint      `gorm:"primarykey" json:"id"`
	TenantID uint      `json:"tenant_id" gorm:"not null;index"`
	Date     time.Time `json:"date" gorm:"not null;uniqueIndex:idx_tenant_daily"`

	// Ringkasan
	TotalOrders   int     `json:"total_orders" gorm:"default:0"`
	TotalRevenue  float64 `json:"total_revenue" gorm:"default:0"`
	TotalExpenses float64 `json:"total_expenses" gorm:"default:0"`
	NetProfit     float64 `json:"net_profit" gorm:"default:0"`

	// Detail per kategori
	CategorySummary JSONMap `json:"category_summary" gorm:"type:jsonb"` // {"food": 500000, "drink": 300000}
	ExpenseSummary  JSONMap `json:"expense_summary" gorm:"type:jsonb"`  // {"operational": 100000, "salary": 500000}

	// Payment methods
	PaymentSummary JSONMap `json:"payment_summary" gorm:"type:jsonb"` // {"cash": 400000, "qris": 400000}

	// Waktu operasional
	FirstOrderTime *time.Time `json:"first_order_time"`
	LastOrderTime  *time.Time `json:"last_order_time"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// JSONMap is a map that can be stored as JSON in database
type JSONMap map[string]interface{}

func (j JSONMap) Value() (interface{}, error) {
	return j, nil
}

func (j *JSONMap) Scan(value interface{}) error {
	return nil
}
