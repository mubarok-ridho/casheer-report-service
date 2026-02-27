package repository

import (
	"github.com/mubarok-ridho/casheer-report-service/internal/models"
	"gorm.io/gorm"
)

type TemplateRepository struct {
	DB *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{DB: db}
}

// Create template
func (r *TemplateRepository) Create(template *models.ReceiptTemplate) error {
	return r.DB.Create(template).Error
}

// Get template by ID
func (r *TemplateRepository) GetByID(id uint, tenantID uint) (*models.ReceiptTemplate, error) {
	var template models.ReceiptTemplate
	err := r.DB.Where("(id = ? AND tenant_id = ?) OR (id = ? AND tenant_id IS NULL)",
		id, tenantID, id).First(&template).Error
	return &template, err
}

// Get all templates for tenant (including system defaults)
func (r *TemplateRepository) GetAll(tenantID uint) ([]models.ReceiptTemplate, error) {
	var templates []models.ReceiptTemplate
	err := r.DB.Where("tenant_id = ? OR tenant_id IS NULL", tenantID).
		Order("is_default desc, name asc").
		Find(&templates).Error
	return templates, err
}

// Get default template
func (r *TemplateRepository) GetDefault(tenantID uint) (*models.ReceiptTemplate, error) {
	var template models.ReceiptTemplate

	// First try tenant-specific default
	err := r.DB.Where("tenant_id = ? AND is_default = ?", tenantID, true).First(&template).Error
	if err == nil {
		return &template, nil
	}

	// Then try system default
	err = r.DB.Where("tenant_id IS NULL AND is_default = ?", true).First(&template).Error
	if err == nil {
		return &template, nil
	}

	// Finally, get any system template
	err = r.DB.Where("tenant_id IS NULL").First(&template).Error
	return &template, err
}

// Update template
func (r *TemplateRepository) Update(template *models.ReceiptTemplate) error {
	return r.DB.Save(template).Error
}

// Delete template
func (r *TemplateRepository) Delete(id uint, tenantID uint) error {
	// Only allow deletion of tenant's own templates
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&models.ReceiptTemplate{}).Error
}

// Set as default template
func (r *TemplateRepository) SetDefault(id uint, tenantID uint) error {
	tx := r.DB.Begin()

	// Remove default from all tenant templates
	if err := tx.Model(&models.ReceiptTemplate{}).
		Where("tenant_id = ?", tenantID).
		Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Set new default
	if err := tx.Model(&models.ReceiptTemplate{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Get templates by paper width
func (r *TemplateRepository) GetByWidth(tenantID uint, width string) ([]models.ReceiptTemplate, error) {
	var templates []models.ReceiptTemplate
	err := r.DB.Where("(tenant_id = ? OR tenant_id IS NULL) AND paper_width = ?", tenantID, width).
		Order("is_default desc").
		Find(&templates).Error
	return templates, err
}
