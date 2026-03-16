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

	r.DB.Where("tenant_id = ? AND date = ?", tenantID, startOfDay).First(&report)
	// Always delete stale cache and recalculate
	r.DB.Where("tenant_id = ? AND date = ?", tenantID, startOfDay).Delete(&models.DailyReport{})
	return r.calculateDailyReport(tenantID, date)
}

// calculateDailyReport - query directly from orders table
func (r *ReportRepository) calculateDailyReport(tenantID uint, date time.Time) (*models.DailyReport, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var revenue float64
	var orderCount int64
	r.DB.Raw("SELECT COALESCE(SUM(total_amount), 0), COUNT(*) FROM orders WHERE tenant_id = ? AND payment_status = 'paid' AND created_at BETWEEN ? AND ?",
		tenantID, startOfDay, endOfDay).Row().Scan(&revenue, &orderCount)

	var expense float64
	r.DB.Raw("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE tenant_id = ? AND date::date = ?::date",
		tenantID, startOfDay).Row().Scan(&expense)

	type PaymentRow struct {
		Method string
		Total  float64
	}
	var paymentRows []PaymentRow
	r.DB.Raw("SELECT payment_method as method, COALESCE(SUM(total_amount), 0) as total FROM orders WHERE tenant_id = ? AND payment_status = 'paid' AND created_at BETWEEN ? AND ? GROUP BY payment_method",
		tenantID, startOfDay, endOfDay).Scan(&paymentRows)
	paymentSummary := models.JSONMap{}
	for _, p := range paymentRows {
		paymentSummary[p.Method] = p.Total
	}

	// Category summary - join orders with order_items and menus
	type CategoryRow struct {
		Category string
		Total    float64
	}
	var categoryRows []CategoryRow
	r.DB.Raw(`SELECT c.name as category, COALESCE(SUM(oi.subtotal), 0) as total
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN menus m ON m.id = oi.menu_id
		JOIN categories c ON c.id = m.category_id
		WHERE o.tenant_id = ? AND o.payment_status = 'paid' AND o.created_at BETWEEN ? AND ?
		GROUP BY c.name`,
		tenantID, startOfDay, endOfDay).Scan(&categoryRows)
	categorySummary := models.JSONMap{}
	for _, c := range categoryRows {
		categorySummary[c.Category] = c.Total
	}

	report := &models.DailyReport{
		TenantID:        tenantID,
		Date:            startOfDay,
		TotalOrders:     int(orderCount),
		TotalRevenue:    revenue,
		TotalExpenses:   expense,
		NetProfit:       revenue - expense,
		PaymentSummary:  paymentSummary,
		CategorySummary: categorySummary,
	}
	// Delete old cached record and save fresh data
	r.DB.Where("tenant_id = ? AND date = ?", tenantID, startOfDay).Delete(&models.DailyReport{})
	r.DB.Create(report)
	return report, nil
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

	type DailyRevenue struct {
		Date    time.Time `gorm:"column:date"`
		Revenue float64   `gorm:"column:revenue"`
	}

	var rows []DailyRevenue
	err := r.DB.Raw(`
		SELECT
			DATE(created_at) as date,
			COALESCE(SUM(total_amount), 0) as revenue
		FROM orders
		WHERE tenant_id = ?
		  AND payment_status = 'paid'
		  AND created_at BETWEEN ? AND ?
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, tenantID, startDate, endDate).Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	var total float64
	var data []map[string]interface{}
	for _, row := range rows {
		total += row.Revenue
		data = append(data, map[string]interface{}{
			"date":    row.Date.Format("2006-01-02"),
			"revenue": row.Revenue,
		})
	}

	average := float64(0)
	if len(rows) > 0 {
		average = total / float64(len(rows))
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
