package request_repo

import (
	"context"
	"errors"
	"fmt"

	"go-init/internal/database"
	dbModel "go-init/internal/database/request_repo/models"

	"github.com/google/uuid"
	orm "gitlab.com/go-init/go-init-common/default/db/pg/orm"
	"gitlab.com/go-init/go-init-common/default/logger"
	"gorm.io/gorm"
)

// ToDo сейчас есть бан при котором в методе получения шпблона по uuid реквеста мтоды пытвется получить сразу шаблон но надо сначала пойти в табличку с реквестом и потом из нее получить id габлона
// Repository implements the GoInitManagerRepository interface
// and provides data access methods for the application
type Repository struct {
	log        *logger.Logger
	schemaName string
	db         *orm.AgentImpl
}

// NewRepository creates a new instance of Repository with direct database access
// If schemaName is empty, "go_init" will be used as default
func NewRepository(db *orm.AgentImpl, log *logger.Logger, schemaName ...string) database.GoInitManagerRepository {
	schema := "go_init" // Default schema name

	// Override default schema if provided
	if len(schemaName) > 0 && schemaName[0] != "" {
		schema = schemaName[0]
	}

	return &Repository{
		log:        log,
		schemaName: schema,
		db:         db,
	}
}

// CreateNewTemplate creates a new template in the database using the provided transaction
func (r *Repository) CreateNewTemplate(ctx context.Context, model *dbModel.ServiceTemplate, tx *orm.Transaction) error {
	return tx.Tx.WithContext(ctx).Create(model).Error
}

// GetTemplateByUUID retrieves a template by its UUID with all related data
func (r *Repository) GetTemplateByUUID(ctx context.Context, templateUUID uuid.UUID) (*dbModel.ServiceTemplate, error) {
	var template dbModel.ServiceTemplate
	err := r.db.DB().WithContext(ctx).
		Preload("Endpoints").
		Preload("DatabaseConfigs").
		Preload("DockerConfigs").
		Preload("AdvancedConfigs").
		Where("service_template_uuid = ?", templateUUID).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("template not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return &template, nil
}

// GetTemplateByID retrieves a template by its numeric ID with all related data
func (r *Repository) GetTemplateByID(ctx context.Context, templateID int) (*dbModel.ServiceTemplate, error) {
	var template dbModel.ServiceTemplate
	err := r.db.DB().WithContext(ctx).
		Preload("Endpoints").
		Preload("DatabaseConfigs").
		Preload("DockerConfigs").
		Preload("AdvancedConfigs").
		Where("service_template_id = ?", templateID).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("template not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return &template, nil
}

// GetRecentTemplates retrieves a list of recent templates ordered by creation date
func (r *Repository) GetRecentTemplates(ctx context.Context, limit int) ([]*dbModel.ServiceTemplate, error) {
	var templates []*dbModel.ServiceTemplate
	err := r.db.DB().WithContext(ctx).
		Preload("Endpoints").
		Preload("DatabaseConfigs").
		Preload("DockerConfigs").
		Preload("AdvancedConfigs").
		Order("created_at DESC").
		Limit(limit).
		Find(&templates).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get recent templates: %w", err)
	}
	return templates, nil
}

// UpdateZipUrl updates the zip url for a template identified by UUID.
func (r *Repository) UpdateZipUrl(ctx context.Context, templateUUID uuid.UUID, newZipUrl string) error {
	var template dbModel.ServiceTemplate
	err := r.db.DB().WithContext(ctx).
		Where("service_template_uuid = ?", templateUUID).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template not found: %w", err)
		}
		return fmt.Errorf("failed to find template: %w", err)
	}

	// Update the zip url
	template.ZipURL = &newZipUrl
	if err := r.db.DB().WithContext(ctx).Save(&template).Error; err != nil {
		return fmt.Errorf("failed to update zip URL: %w", err)
	}

	return nil
}

// UpdateTemplateStatusByUUID updates the status of a template identified by UUID.
func (r *Repository) UpdateTemplateStatusByUUID(ctx context.Context, templateUUID uuid.UUID, newStatus string) error {
	var template dbModel.ServiceTemplate
	err := r.db.DB().WithContext(ctx).
		Where("service_template_uuid = ?", templateUUID).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template not found: %w", err)
		}
		return fmt.Errorf("failed to find template: %w", err)
	}

	template.Status = &newStatus
	if err := r.db.DB().WithContext(ctx).Save(&template).Error; err != nil {
		return fmt.Errorf("failed to update template status: %w", err)
	}

	return nil
}

// UpdateTemplateErrorByUUID updates the error message of a template identified by UUID.
func (r *Repository) UpdateTemplateErrorByUUID(ctx context.Context, templateUUID uuid.UUID, errorMessage string) error {
	var template dbModel.ServiceTemplate
	err := r.db.DB().WithContext(ctx).
		Where("service_template_uuid = ?", templateUUID).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template not found: %w", err)
		}
		return fmt.Errorf("failed to find template: %w", err)
	}

	template.Error = &errorMessage
	if err := r.db.DB().WithContext(ctx).Save(&template).Error; err != nil {
		return fmt.Errorf("failed to update template error message: %w", err)
	}

	return nil
}
