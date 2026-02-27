package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mubarok-ridho/casheer-report-service/internal/models"
	"github.com/mubarok-ridho/casheer-report-service/internal/repository"
	"github.com/mubarok-ridho/casheer-report-service/pkg/messaging"
)

type ExpenseHandler struct {
	repo *repository.ExpenseRepository
	rmq  *messaging.RabbitMQ
}

func NewExpenseHandler(repo *repository.ExpenseRepository) *ExpenseHandler {
	rmq, _ := messaging.NewRabbitMQ()
	return &ExpenseHandler{
		repo: repo,
		rmq:  rmq,
	}
}

// Create expense
func (h *ExpenseHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	var input struct {
		Category      string  `json:"category"`
		Description   string  `json:"description"`
		Amount        float64 `json:"amount"`
		Date          string  `json:"date"`
		PaymentMethod string  `json:"payment_method"`
		Notes         string  `json:"notes"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validasi
	if input.Category == "" || input.Description == "" || input.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Category, description, and valid amount are required",
		})
	}

	// Parse date
	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		date = time.Now()
	}

	expense := &models.Expense{
		TenantID:      tenantID,
		Category:      input.Category,
		Description:   input.Description,
		Amount:        input.Amount,
		Date:          date,
		PaymentMethod: input.PaymentMethod,
		Notes:         input.Notes,
	}

	// Handle receipt upload if any
	file, _ := c.FormFile("receipt")
	if file != nil {
		// Upload receipt to cloud storage
		// receiptURL, _ := utils.UploadReceipt(file, tenantID)
		// expense.ReceiptURL = receiptURL
	}

	if err := h.repo.Create(expense); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Publish event ke RabbitMQ
	if h.rmq != nil {
		event := messaging.ExpenseAddedEvent{
			ExpenseID: expense.ID,
			TenantID:  tenantID,
			Amount:    expense.Amount,
			Category:  expense.Category,
			Date:      expense.Date.Format("2006-01-02"),
		}
		h.rmq.PublishExpenseAdded(event)
	}

	return c.Status(201).JSON(expense)
}

// Get all expenses
func (h *ExpenseHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	category := c.Query("category")

	expenses, total, err := h.repo.GetAll(tenantID, page, limit, startDate, endDate, category)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":  expenses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Get expense by ID
func (h *ExpenseHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid expense ID"})
	}

	expense, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Expense not found"})
	}

	return c.JSON(expense)
}

// Update expense
func (h *ExpenseHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid expense ID"})
	}

	var input struct {
		Category      string  `json:"category"`
		Description   string  `json:"description"`
		Amount        float64 `json:"amount"`
		Date          string  `json:"date"`
		PaymentMethod string  `json:"payment_method"`
		Notes         string  `json:"notes"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	expense, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Expense not found"})
	}

	date, _ := time.Parse("2006-01-02", input.Date)

	expense.Category = input.Category
	expense.Description = input.Description
	expense.Amount = input.Amount
	expense.Date = date
	expense.PaymentMethod = input.PaymentMethod
	expense.Notes = input.Notes

	if err := h.repo.Update(expense); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(expense)
}

// Delete expense
func (h *ExpenseHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid expense ID"})
	}

	if err := h.repo.Delete(uint(id), tenantID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Expense deleted successfully"})
}

// Get by category summary
func (h *ExpenseHandler) GetByCategory(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	summary, err := h.repo.GetByCategory(tenantID, startDate, endDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(summary)
}
