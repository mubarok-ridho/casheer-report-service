package repository

import (
	"time"

	"github.com/mubarok-ridho/casheer-report-service/internal/models"
	"gorm.io/gorm"
)

type ReportRepository struct {
	DB *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{DB: db}
}

// Get daily report
func (r *ReportRepository) GetDailyReport(tenantID uint, date time.Time) (*models.DailyReport, error) {
	var report models.DailyReport
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	err := r.DB.Where("tenant_id = ? AND date = ?", tenantID, startOfDay).First(&report).Error
	if err != nil {
		// If not found, calculate on the fly
		return r.calculateDailyReport(tenantID, date)
	}
	return &report, nil
}

// Get monthly report
func (r *ReportRepository) GetMonthlyReport(tenantID uint, month, year int) (map[string]interface{}, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	var results []struct {
		Date       string  `json:"date"`
		Revenue    float64 `json:"revenue"`
		Expense    float64 `json:"expense"`
		NetProfit  float64 `json:"net_profit"`
		OrderCount int     `json:"order_count"`
	}

	err := r.DB.Table("daily_reports").
		Select("date, total_revenue as revenue, total_expenses as expense, net_profit, total_orders as order_count").
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Order("date asc").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Calculate totals
	var totalRevenue, totalExpense, totalProfit float64
	var totalOrders int

	for _, r := range results {
		totalRevenue += r.Revenue
		totalExpense += r.Expense
		totalProfit += r.NetProfit
		totalOrders += r.OrderCount
	}

	return map[string]interface{}{
		"daily":         results,
		"total_revenue": totalRevenue,
		"total_expense": totalExpense,
		"total_profit":  totalProfit,
		"total_orders":  totalOrders,
		"month":         month,
		"year":          year,
	}, nil
}

// Get yearly report
func (r *ReportRepository) GetYearlyReport(tenantID uint, year int) (map[string]interface{}, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(1, 0, 0)

	var results []struct {
		Month      int     `json:"month"`
		Revenue    float64 `json:"revenue"`
		Expense    float64 `json:"expense"`
		NetProfit  float64 `json:"net_profit"`
		OrderCount int     `json:"order_count"`
	}

	err := r.DB.Table("daily_reports").
		Select("EXTRACT(MONTH FROM date) as month, SUM(total_revenue) as revenue, SUM(total_expenses) as expense, SUM(net_profit) as net_profit, SUM(total_orders) as order_count").
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Group("month").
		Order("month asc").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Calculate totals
	var totalRevenue, totalExpense, totalProfit float64
	var totalOrders int

	for _, r := range results {
		totalRevenue += r.Revenue
		totalExpense += r.Expense
		totalProfit += r.NetProfit
		totalOrders += r.OrderCount
	}

	return map[string]interface{}{
		"monthly":       results,
		"total_revenue": totalRevenue,
		"total_expense": totalExpense,
		"total_profit":  totalProfit,
		"total_orders":  totalOrders,
		"year":          year,
	}, nil
}

// Get revenue summary for last N days
func (r *ReportRepository) GetRevenueSummary(tenantID uint, days int) (map[string]interface{}, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	var revenues []models.Revenue
	err := r.DB.Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Order("date asc").
		Find(&revenues).Error

	if err != nil {
		return nil, err
	}

	var total float64
	var data []map[string]interface{}

	for _, r := range revenues {
		total += r.TotalRevenue
		data = append(data, map[string]interface{}{
			"date":    r.Date.Format("2006-01-02"),
			"revenue": r.TotalRevenue,
		})
	}

	average := total / float64(len(revenues))
	if len(revenues) == 0 {
		average = 0
	}

	return map[string]interface{}{
		"data":    data,
		"total":   total,
		"average": average,
		"days":    days,
	}, nil
}

// Get expense summary for last N days
func (r *ReportRepository) GetExpenseSummary(tenantID uint, days int) (map[string]interface{}, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	var expenses []models.Expense
	err := r.DB.Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Order("date asc").
		Find(&expenses).Error

	if err != nil {
		return nil, err
	}

	var total float64
	var byCategory = make(map[string]float64)
	var data []map[string]interface{}

	for _, e := range expenses {
		total += e.Amount
		byCategory[e.Category] += e.Amount
		data = append(data, map[string]interface{}{
			"date":        e.Date.Format("2006-01-02"),
			"amount":      e.Amount,
			"category":    e.Category,
			"description": e.Description,
		})
	}

	return map[string]interface{}{
		"data":        data,
		"total":       total,
		"by_category": byCategory,
		"days":        days,
	}, nil
}

// calculateDailyReport - internal function to calculate report on the fly
func (r *ReportRepository) calculateDailyReport(tenantID uint, date time.Time) (*models.DailyReport, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get revenue from events
	var revenue float64
	var orderCount int64
	r.DB.Model(&models.RevenueEvent{}).
		Where("tenant_id = ? AND type = 'order' AND date BETWEEN ? AND ?",
			tenantID, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0), COUNT(*)").
		Row().Scan(&revenue, &orderCount)

	// Get expenses
	var expense float64
	var expenseCount int64
	r.DB.Model(&models.Expense{}).
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0), COUNT(*)").
		Row().Scan(&expense, &expenseCount)

	report := &models.DailyReport{
		TenantID:      tenantID,
		Date:          startOfDay,
		TotalOrders:   int(orderCount),
		TotalRevenue:  revenue,
		TotalExpenses: expense,
		NetProfit:     revenue - expense,
	}

	// Save to database
	r.DB.Create(report)

	return report, nil
}
