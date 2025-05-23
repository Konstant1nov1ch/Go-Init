package database

import (
	"context"

	dbModel "go-init/internal/database/request_repo/models"

	"github.com/google/uuid"
	"gitlab.com/go-init/go-init-common/default/db/pg/orm"
)

type GoInitManagerRepository interface {
	CreateNewTemplate(ctx context.Context, model *dbModel.ServiceTemplate, tx *orm.Transaction) error
	GetTemplateByUUID(ctx context.Context, templateUUID uuid.UUID) (*dbModel.ServiceTemplate, error)
	GetTemplateByID(ctx context.Context, templateID int) (*dbModel.ServiceTemplate, error)
	GetRecentTemplates(ctx context.Context, limit int) ([]*dbModel.ServiceTemplate, error)
	UpdateZipUrl(ctx context.Context, templateUUID uuid.UUID, newZipUrl string) error
	UpdateTemplateStatusByUUID(ctx context.Context, templateUUID uuid.UUID, newStatus string) error
	UpdateTemplateErrorByUUID(ctx context.Context, templateUUID uuid.UUID, errorMessage string) error

	// ...
}
