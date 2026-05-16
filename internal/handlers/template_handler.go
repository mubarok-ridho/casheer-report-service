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
	return &TemplateHandler{repo: repo, db: db}
}

// input struct lengkap — dipakai untuk Create & Update
type templateInput struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Header           string `json:"header"`
	Footer           string `json:"footer"`
	ShowLogo         bool   `json:"show_logo"`
	ShowTax          bool   `json:"show_tax"`
	ShowDiscount     bool   `json:"show_discount"`
	ShowVariations   bool   `json:"show_variations"`
	ShowNotes        bool   `json:"show_notes"`
	PaperWidth       string `json:"paper_width"`
	FontSize         int    `json:"font_size"`
	LogoPosition     string `json:"logo_position"`
	MarginTop        int    `json:"margin_top"`
	MarginBottom     int    `json:"margin_bottom"`
	IsActive         bool   `json:"is_active"`
}

func (h *TemplateHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	var input templateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if input.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Template name is required"})
	}
	if input.PaperWidth == "" { input.PaperWidth = "58mm" }
	if input.FontSize == 0 { input.FontSize = 12 }
	if input.LogoPosition == "" { input.LogoPosition = "center" }

	charsPerLine := 32
	if input.PaperWidth == "80mm" { charsPerLine = 48 }

	template := &models.ReceiptTemplate{
		TenantID:          &tenantID,
		Name:              input.Name,
		Description:       input.Description,
		Header:            input.Header,
		Footer:            input.Footer,
		ShowLogo:          input.ShowLogo,
		ShowTax:           input.ShowTax,
		ShowDiscount:      input.ShowDiscount,
		ShowVariations:    input.ShowVariations,
		ShowNotes:         input.ShowNotes,
		PaperWidth:        input.PaperWidth,
		FontSize:          input.FontSize,
		CharactersPerLine: charsPerLine,
		LogoPosition:      input.LogoPosition,
		MarginTop:         input.MarginTop,
		MarginBottom:      input.MarginBottom,
		IsDefault:         false,
		IsActive:          true,
	}

	if err := h.repo.Create(template); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(template)
}

func (h *TemplateHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	templates, err := h.repo.GetAll(tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(templates)
}

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

func (h *TemplateHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid template ID"})
	}

	var input templateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	template, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Template not found"})
	}

	if input.LogoPosition == "" { input.LogoPosition = "center" }
	charsPerLine := 32
	if input.PaperWidth == "80mm" { charsPerLine = 48 }

	template.Name             = input.Name
	template.Description      = input.Description
	template.Header           = input.Header
	template.Footer           = input.Footer
	template.ShowLogo         = input.ShowLogo
	template.ShowTax          = input.ShowTax
	template.ShowDiscount     = input.ShowDiscount
	template.ShowVariations   = input.ShowVariations
	template.ShowNotes        = input.ShowNotes
	template.PaperWidth       = input.PaperWidth
	template.FontSize         = input.FontSize
	template.CharactersPerLine = charsPerLine
	template.LogoPosition     = input.LogoPosition
	template.MarginTop        = input.MarginTop
	template.MarginBottom     = input.MarginBottom
	template.IsActive         = input.IsActive

	if err := h.repo.Update(template); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(template)
}

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
	if input.Copies <= 0 { input.Copies = 1 }

	var template *models.ReceiptTemplate
	if input.TemplateID != nil {
		template, err = h.repo.GetByID(*input.TemplateID, tenantID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Template not found"})
		}
	} else {
		template, err = h.repo.GetDefault(tenantID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "No default template found"})
		}
	}

	orderData := map[string]interface{}{
		"order_id": orderID, "order_number": "ORD-20240227-001",
		"date": time.Now().Format("2006-01-02 15:04:05"), "customer": "Walk-in Customer",
		"items": []map[string]interface{}{
			{"name": "Nasi Goreng", "quantity": 2, "price": 25000, "subtotal": 50000},
		},
		"subtotal": 50000, "tax": 0, "total": 50000, "payment": "Tunai",
	}
	tenantInfo := map[string]interface{}{"name": "Toko", "address": "", "phone": ""}

	receiptData, err := utils.GenerateReceipt(tenantInfo, orderData, template)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate receipt: " + err.Error()})
	}
	for i := 0; i < input.Copies; i++ {
		if err := utils.PrintToBluetooth(input.PrinterMAC, receiptData); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to print: " + err.Error(), "copy": i + 1})
		}
	}
	return c.JSON(fiber.Map{"message": "Receipt printed successfully", "copies": input.Copies})
}

func (h *TemplateHandler) PrintTest(c *fiber.Ctx) error {
	var input struct {
		PrinterMAC string `json:"printer_mac"`
		PaperWidth string `json:"paper_width"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	testData := map[string]interface{}{
		"title": "TEST PRINT", "date": time.Now().Format("2006-01-02 15:04:05"),
		"content": "Printer bekerja dengan baik", "width": input.PaperWidth,
	}
	testReceipt := utils.GenerateTestPage(testData)
	if err := utils.PrintToBluetooth(input.PrinterMAC, testReceipt); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to print: " + err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Test page printed successfully"})
}
