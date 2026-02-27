package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mubarok-ridho/casheer-report-service/internal/models"
	"github.com/mubarok-ridho/casheer-report-service/internal/repository"
	"github.com/mubarok-ridho/casheer-report-service/internal/utils"
	"gorm.io/gorm"
)

type TemplateHandler struct {
	repo *repository.TemplateRepository
	db   *gorm.DB
}

func NewTemplateHandler(repo *repository.TemplateRepository, db *gorm.DB) *TemplateHandler {
	return &TemplateHandler{
		repo: repo,
		db:   db,
	}
}

// Create template
func (h *TemplateHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	var input struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Header       string `json:"header"`
		Footer       string `json:"footer"`
		ShowLogo     bool   `json:"show_logo"`
		ShowTax      bool   `json:"show_tax"`
		ShowDiscount bool   `json:"show_discount"`
		PaperWidth   string `json:"paper_width"`
		FontSize     int    `json:"font_size"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if input.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Template name is required"})
	}

	// Set default values
	if input.PaperWidth == "" {
		input.PaperWidth = "58mm"
	}
	if input.FontSize == 0 {
		input.FontSize = 12
	}

	charsPerLine := 32
	if input.PaperWidth == "80mm" {
		charsPerLine = 48
	}

	template := &models.ReceiptTemplate{
		TenantID:          &tenantID,
		Name:              input.Name,
		Description:       input.Description,
		Header:            input.Header,
		Footer:            input.Footer,
		ShowLogo:          input.ShowLogo,
		ShowTax:           input.ShowTax,
		ShowDiscount:      input.ShowDiscount,
		PaperWidth:        input.PaperWidth,
		FontSize:          input.FontSize,
		CharactersPerLine: charsPerLine,
		IsDefault:         false,
		IsActive:          true,
	}

	if err := h.repo.Create(template); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(template)
}

// Get all templates
func (h *TemplateHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	templates, err := h.repo.GetAll(tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(templates)
}

// Get template by ID
func (h *TemplateHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid template ID"})
	}

	template, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Template not found"})
	}

	return c.JSON(template)
}

// Update template
func (h *TemplateHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid template ID"})
	}

	var input struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Header       string `json:"header"`
		Footer       string `json:"footer"`
		ShowLogo     bool   `json:"show_logo"`
		ShowTax      bool   `json:"show_tax"`
		ShowDiscount bool   `json:"show_discount"`
		PaperWidth   string `json:"paper_width"`
		FontSize     int    `json:"font_size"`
		IsActive     bool   `json:"is_active"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	template, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Template not found"})
	}

	charsPerLine := 32
	if input.PaperWidth == "80mm" {
		charsPerLine = 48
	}

	template.Name = input.Name
	template.Description = input.Description
	template.Header = input.Header
	template.Footer = input.Footer
	template.ShowLogo = input.ShowLogo
	template.ShowTax = input.ShowTax
	template.ShowDiscount = input.ShowDiscount
	template.PaperWidth = input.PaperWidth
	template.FontSize = input.FontSize
	template.CharactersPerLine = charsPerLine
	template.IsActive = input.IsActive

	if err := h.repo.Update(template); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(template)
}

// Delete template
func (h *TemplateHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid template ID"})
	}

	if err := h.repo.Delete(uint(id), tenantID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Template deleted successfully"})
}

// Set as default template
func (h *TemplateHandler) SetDefault(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid template ID"})
	}

	if err := h.repo.SetDefault(uint(id), tenantID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Template set as default successfully"})
}

// Print receipt
func (h *TemplateHandler) PrintReceipt(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	orderID, err := strconv.ParseUint(c.Params("orderId"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid order ID"})
	}

	var input struct {
		TemplateID *uint  `json:"template_id"`
		PrinterMAC string `json:"printer_mac"`
		Copies     int    `json:"copies"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if input.Copies <= 0 {
		input.Copies = 1
	}

	// Get template
	var template *models.ReceiptTemplate
	if input.TemplateID != nil {
		template, err = h.repo.GetByID(*input.TemplateID, tenantID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Template not found"})
		}
	} else {
		// Get default template
		template, err = h.repo.GetDefault(tenantID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "No default template found"})
		}
	}

	// Get order data from menu service (via API call)
	// For now, we'll simulate order data
	orderData := map[string]interface{}{
		"order_id":     orderID,
		"order_number": "ORD-20240227-001",
		"date":         time.Now().Format("2006-01-02 15:04:05"),
		"customer":     "Walk-in Customer",
		"items": []map[string]interface{}{
			{
				"name":     "Nasi Goreng",
				"quantity": 2,
				"price":    25000,
				"subtotal": 50000,
			},
			{
				"name":     "Es Teh",
				"quantity": 2,
				"price":    5000,
				"subtotal": 10000,
			},
		},
		"subtotal": 60000,
		"tax":      3000,
		"total":    63000,
		"payment":  "Tunai",
	}

	// Get tenant info from auth service (via API call)
	tenantInfo := map[string]interface{}{
		"name":    "Warung Sederhana",
		"address": "Jl. Raya No. 123",
		"phone":   "08123456789",
	}

	// Generate receipt
	receiptData, err := utils.GenerateReceipt(tenantInfo, orderData, template)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate receipt: " + err.Error()})
	}

	// Send to printer
	for i := 0; i < input.Copies; i++ {
		if err := utils.PrintToBluetooth(input.PrinterMAC, receiptData); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to print: " + err.Error(),
				"copy":  i + 1,
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Receipt printed successfully",
		"copies":  input.Copies,
	})
}

// Print test page
func (h *TemplateHandler) PrintTest(c *fiber.Ctx) error {
	var input struct {
		PrinterMAC string `json:"printer_mac"`
		PaperWidth string `json:"paper_width"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Create test data
	testData := map[string]interface{}{
		"title":   "TEST PRINT",
		"date":    time.Now().Format("2006-01-02 15:04:05"),
		"content": "Printer bekerja dengan baik",
		"width":   input.PaperWidth,
	}

	testReceipt := utils.GenerateTestPage(testData)

	if err := utils.PrintToBluetooth(input.PrinterMAC, testReceipt); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to print: " + err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Test page printed successfully"})
}
