package handlers

import (
	"casheer-report-service/internal/models"
	"casheer-report-service/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TemplateHandler struct {
	DB *gorm.DB
}

func NewTemplateHandler(db *gorm.DB) *TemplateHandler {
	return &TemplateHandler{DB: db}
}

// Get all templates
func (h *TemplateHandler) GetTemplates(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	var templates []models.ReceiptTemplate
	h.DB.Where("tenant_id = ? OR is_default = ?", tenantID, true).Find(&templates)

	return c.JSON(templates)
}

// Create custom template
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	var input struct {
		Name       string `json:"name"`
		Header     string `json:"header"`
		Footer     string `json:"footer"`
		ShowLogo   bool   `json:"show_logo"`
		ShowTax    bool   `json:"show_tax"`
		PaperWidth string `json:"paper_width"` // "58mm" atau "80mm"
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	template := models.ReceiptTemplate{
		TenantID:   &tenantID,
		Name:       input.Name,
		Header:     input.Header,
		Footer:     input.Footer,
		ShowLogo:   input.ShowLogo,
		ShowTax:    input.ShowTax,
		PaperWidth: input.PaperWidth,
		IsDefault:  false,
	}

	if err := h.DB.Create(&template).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(template)
}

// Print receipt (dengan bluetooth)
func (h *TemplateHandler) PrintReceipt(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	var input struct {
		OrderID    uint   `json:"order_id"`
		TemplateID *uint  `json:"template_id"`
		PrinterMAC string `json:"printer_mac"` // Alamat MAC printer bluetooth
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Get order data (panggil menu service)
	// Untuk contoh, kita asumsikan sudah dapat data order

	// Get template
	var template models.ReceiptTemplate
	if input.TemplateID != nil {
		h.DB.First(&template, input.TemplateID)
	} else {
		// Get default template untuk tenant
		h.DB.Where("tenant_id = ? OR is_default = ?", tenantID, true).
			Order("is_default DESC").
			First(&template)
	}

	// Get tenant info
	var tenant struct {
		StoreName string
		LogoURL   string
	}
	// Ambil dari auth service via API atau dari local cache

	// Generate receipt
	receiptData := utils.GenerateReceipt(tenant, template, input.OrderID)

	// Print ke bluetooth printer
	err := utils.PrintToBluetooth(input.PrinterMAC, receiptData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to print: " + err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Print job sent successfully"})
}
