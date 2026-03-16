package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mubarok-ridho/casheer-report-service/internal/models"
	"github.com/mubarok-ridho/casheer-report-service/internal/repository"
	"github.com/mubarok-ridho/casheer-report-service/pkg/messaging"
	"gorm.io/gorm"
)

type ReportHandler struct {
	db         *gorm.DB
	reportRepo *repository.ReportRepository
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{
		db:         db,
		reportRepo: repository.NewReportRepository(db),
	}
}

// HandleOrderCompleted - dipanggil dari RabbitMQ consumer
func (h *ReportHandler) HandleOrderCompleted(event messaging.OrderCompletedEvent) {
	date, _ := time.Parse("2006-01-02", event.Date)

	// Update atau create revenue
	var revenue models.Revenue
	result := h.db.Where("tenant_id = ? AND date = ?", event.TenantID, date).First(&revenue)

	if result.Error != nil {
		// Create new
		revenue = models.Revenue{
			TenantID:     event.TenantID,
			Date:         date,
			TotalRevenue: event.TotalAmount,
			TotalExpense: 0,
			NetRevenue:   event.TotalAmount,
			OrderCount:   1,
			ExpenseCount: 0,
		}
		h.db.Create(&revenue)
	} else {
		// Update existing
		revenue.TotalRevenue += event.TotalAmount
		revenue.NetRevenue = revenue.TotalRevenue - revenue.TotalExpense
		revenue.OrderCount += 1
		h.db.Save(&revenue)
	}

	// Record event
	h.db.Create(&models.RevenueEvent{
		TenantID:    event.TenantID,
		Type:        "order",
		ReferenceID: event.OrderID,
		Amount:      event.TotalAmount,
		Date:        date,
	})

	// Update daily report
	h.updateDailyReport(event.TenantID, date)
}

// GetDailyReport
func (h *ReportHandler) GetDailyReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	dateStr := c.Query("date")
	var date time.Time
	var err error

	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD"})
		}
	}

	report, err := h.reportRepo.GetDailyReport(tenantID, date)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}

// GetMonthlyReport
func (h *ReportHandler) GetMonthlyReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if month == 0 {
		month = int(time.Now().Month())
	}
	if year == 0 {
		year = time.Now().Year()
	}

	report, err := h.reportRepo.GetMonthlyReport(tenantID, month, year)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}

// GetYearlyReport
func (h *ReportHandler) GetYearlyReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	yearStr := c.Query("year")
	year, _ := strconv.Atoi(yearStr)
	if year == 0 {
		year = time.Now().Year()
	}

	report, err := h.reportRepo.GetYearlyReport(tenantID, year)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}

// GetRevenueSummary
func (h *ReportHandler) GetRevenueSummary(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	days := c.QueryInt("days", 30)

	summary, err := h.reportRepo.GetRevenueSummary(tenantID, days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(summary)
}

// GetExpenseSummary
func (h *ReportHandler) GetExpenseSummary(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	days := c.QueryInt("days", 30)

	summary, err := h.reportRepo.GetExpenseSummary(tenantID, days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(summary)
}

// updateDailyReport - internal function
func (h *ReportHandler) updateDailyReport(tenantID uint, date time.Time) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Hitung total orders dan revenue
	var totalRevenue float64
	var orderCount int64
	h.db.Model(&models.RevenueEvent{}).
		Where("tenant_id = ? AND type = 'order' AND date BETWEEN ? AND ?",
			tenantID, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0), COUNT(*)").
		Row().Scan(&totalRevenue, &orderCount)

	// Hitung total expenses
	var totalExpense float64
	var expenseCount int64
	h.db.Model(&models.Expense{}).
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0), COUNT(*)").
		Row().Scan(&totalExpense, &expenseCount)

	// Update atau create daily report
	var report models.DailyReport
	result := h.db.Where("tenant_id = ? AND date = ?", tenantID, startOfDay).First(&report)

	if result.Error != nil {
		report = models.DailyReport{
			TenantID:      tenantID,
			Date:          startOfDay,
			TotalOrders:   int(orderCount),
			TotalRevenue:  totalRevenue,
			TotalExpenses: totalExpense,
			NetProfit:     totalRevenue - totalExpense,
		}
		h.db.Create(&report)
	} else {
		report.TotalOrders = int(orderCount)
		report.TotalRevenue = totalRevenue
		report.TotalExpenses = totalExpense
		report.NetProfit = totalRevenue - totalExpense
		h.db.Save(&report)
	}
}
