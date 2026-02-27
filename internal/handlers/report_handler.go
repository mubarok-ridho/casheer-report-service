package handlers

import (
	"casheer-report-service/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ReportHandler struct {
	DB *gorm.DB
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{DB: db}
}

// Get daily report
func (h *ReportHandler) GetDailyReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	dateStr := c.Query("date")
	var date time.Time

	if dateStr == "" {
		date = time.Now()
	} else {
		date, _ = time.Parse("2006-01-02", dateStr)
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get total revenue from orders (via API call ke menu service atau event)
	// Untuk contoh, kita hitung dari tabel revenue yang sudah di-populate via event

	var revenue models.Revenue
	result := h.DB.Where("tenant_id = ? AND date = ?", tenantID, startOfDay).First(&revenue)

	if result.Error != nil {
		// Jika belum ada, hitung manual atau return 0
		revenue = models.Revenue{
			TenantID:     tenantID,
			Date:         startOfDay,
			TotalRevenue: 0,
			TotalExpense: 0,
			NetRevenue:   0,
		}
	}

	// Get expenses hari ini
	var expenses []models.Expense
	h.DB.Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startOfDay, endOfDay).
		Find(&expenses)

	return c.JSON(fiber.Map{
		"date":     date.Format("2006-01-02"),
		"revenue":  revenue,
		"expenses": expenses,
	})
}

// Add expense
func (h *ReportHandler) AddExpense(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	var input struct {
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Amount      float64 `json:"amount"`
		Date        string  `json:"date"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		date = time.Now()
	}

	expense := models.Expense{
		TenantID:    tenantID,
		Category:    input.Category,
		Description: input.Description,
		Amount:      input.Amount,
		Date:        date,
	}

	if err := h.DB.Create(&expense).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Update atau buat revenue record untuk hari tersebut
	h.updateRevenue(tenantID, date)

	return c.Status(201).JSON(expense)
}

// Update revenue (dipanggil setelah ada order atau expense baru)
func (h *ReportHandler) updateRevenue(tenantID uint, date time.Time) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	// Hitung total revenue dari orders (via event table atau langsung panggil menu service)
	// Untuk contoh, kita pakai aggregate dari tabel revenue_events

	var totalRevenue float64
	h.DB.Table("revenue_events").
		Where("tenant_id = ? AND date = ? AND type = 'order'", tenantID, startOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRevenue)

	var totalExpense float64
	h.DB.Model(&models.Expense{}).
		Where("tenant_id = ? AND date = ?", tenantID, startOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense)

	// Upsert revenue
	var revenue models.Revenue
	result := h.DB.Where("tenant_id = ? AND date = ?", tenantID, startOfDay).First(&revenue)

	if result.Error != nil {
		revenue = models.Revenue{
			TenantID:     tenantID,
			Date:         startOfDay,
			TotalRevenue: totalRevenue,
			TotalExpense: totalExpense,
			NetRevenue:   totalRevenue - totalExpense,
		}
		h.DB.Create(&revenue)
	} else {
		revenue.TotalRevenue = totalRevenue
		revenue.TotalExpense = totalExpense
		revenue.NetRevenue = totalRevenue - totalExpense
		h.DB.Save(&revenue)
	}
}
