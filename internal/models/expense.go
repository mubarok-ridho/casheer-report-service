package models

import (
    "time"
    "gorm.io/gorm"
)

type Expense struct {
    ID          uint           `gorm:"primarykey" json:"id"`
    TenantID    uint           `json:"tenant_id"`
    Category    string         `json:"category"` // "Operational", "Marketing", dll
    Description string         `json:"description"`
    Amount      float64        `json:"amount"`
    Date        time.Time      `json:"date"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Revenue struct {
    ID            uint      `gorm:"primarykey" json:"id"`
    TenantID      uint      `json:"tenant_id"`
    Date          time.Time `json:"date"`
    TotalRevenue  float64   `json:"total_revenue"`
    TotalExpense  float64   `json:"total_expense"`
    NetRevenue    float64   `json:"net_revenue"`
    CreatedAt     time.Time `json:"created_at"`
}