package repository

import (
	"time"

	"github.com/mubarok-ridho/casheer-report-service/internal/models"
	"gorm.io/gorm"
)

type ExpenseRepository struct {
	DB *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{DB: db}
}

// Create expense
func (r *ExpenseRepository) Create(expense *models.Expense) error {
	return r.DB.Create(expense).Error
}

// Get expense by ID
func (r *ExpenseRepository) GetByID(id uint, tenantID uint) (*models.Expense, error) {
	var expense models.Expense
	err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&expense).Error
	return &expense, err
}

// Get all expenses with pagination
func (r *ExpenseRepository) GetAll(tenantID uint, page, limit int, startDate, endDate, category string) ([]models.Expense, int64, error) {
	var expenses []models.Expense
	var total int64

	query := r.DB.Model(&models.Expense{}).Where("tenant_id = ?", tenantID)

	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("date desc, created_at desc").
		Offset(offset).
		Limit(limit).
		Find(&expenses).Error

	return expenses, total, err
}

// Update expense
func (r *ExpenseRepository) Update(expense *models.Expense) error {
	return r.DB.Save(expense).Error
}

// Delete expense
func (r *ExpenseRepository) Delete(id uint, tenantID uint) error {
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&models.Expense{}).Error
}

// Get by category summary
func (r *ExpenseRepository) GetByCategory(tenantID uint, startDate, endDate string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := r.DB.Table("expenses").
		Select("category, COUNT(*) as count, SUM(amount) as total").
		Where("tenant_id = ?", tenantID)

	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	err := query.Group("category").Order("total desc").Scan(&results).Error
	return results, err
}

// Get monthly total
func (r *ExpenseRepository) GetMonthlyTotal(tenantID uint, year, month int) (float64, error) {
	var total float64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	err := r.DB.Model(&models.Expense{}).
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error

	return total, err
}
