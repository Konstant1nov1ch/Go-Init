package models

import (
	"time"

	"github.com/google/uuid"
)

// Массив моделей для AutoMigrate в GORM
var Models = []interface{}{
	&ServiceTemplate{},
	&Endpoint{},
	&DatabaseConfig{},
	&DockerConfig{},
	&AdvancedConfig{},
}

// ===========================
// Request
// ===========================
type ServiceTemplate struct {
	ServiceTemplateId   *int       `gorm:"column:service_template_id;primaryKey;autoIncrement"`
	ServiceTemplateUuid *uuid.UUID `gorm:"column:service_template_uuid;type:uuid;default:gen_random_uuid()"`

	ServiceTemplateName *string `gorm:"type:varchar(255);not null"`
	ZipURL              *string `gorm:"type:text;not null"`

	// владелец (кто создал этот шаблон)
	UserId *uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	// Новый статус
	Status *string `gorm:"type:varchar(50);not null;default:'pending'"`
	// Информация об ошибке, если статус FAILED
	Error *string `gorm:"type:text"`

	// Версия API, которая была использована при создании шаблона
	Version *string `gorm:"type:varchar(10)"`

	CreatedAt *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	Endpoints       []*Endpoint       `gorm:"foreignKey:TemplateId;references:ServiceTemplateId;constraint:OnDelete:CASCADE"`
	DatabaseConfigs []*DatabaseConfig `gorm:"foreignKey:TemplateId;references:ServiceTemplateId;constraint:OnDelete:CASCADE"`
	DockerConfigs   []*DockerConfig   `gorm:"foreignKey:TemplateId;references:ServiceTemplateId;constraint:OnDelete:CASCADE"`
	AdvancedConfigs []*AdvancedConfig `gorm:"foreignKey:TemplateId;references:ServiceTemplateId;constraint:OnDelete:CASCADE"`
	// Удалили поле Requests []*Request
}

// ===========================
// Endpoint
// ===========================
type Endpoint struct {
	EndpointId   *int       `gorm:"column:endpoint_id;primaryKey;autoIncrement"`
	EndpointUuid *uuid.UUID `gorm:"column:endpoint_uuid;type:uuid;default:gen_random_uuid()"`

	TemplateId int `gorm:"column:template_id;not null"`

	Protocol  *string    `gorm:"type:varchar(10);not null"` // 'GRPC','REST','GRAPHQL'
	Role      *string    `gorm:"type:varchar(10);not null"` // 'CLIENT','SERVER'
	CreatedAt *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	Template *ServiceTemplate `gorm:"foreignKey:TemplateId;references:ServiceTemplateId;constraint:OnDelete:CASCADE"`
}

// ===========================
// DatabaseConfig
// ===========================
type DatabaseConfig struct {
	DatabaseConfigId   *int       `gorm:"column:database_config_id;primaryKey;autoIncrement"`
	DatabaseConfigUuid *uuid.UUID `gorm:"column:database_config_uuid;type:uuid;default:gen_random_uuid()"`

	TemplateId int `gorm:"column:template_id;not null"`

	Type      *string    `gorm:"type:varchar(10);not null"` // 'POSTGRESQL', 'MYSQL', 'NONE'
	DDL       *string    `gorm:"type:text"`
	CreatedAt *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	Template *ServiceTemplate `gorm:"foreignKey:TemplateId;references:ServiceTemplateId;constraint:OnDelete:CASCADE"`
}

// ===========================
// DockerConfig
// ===========================
type DockerConfig struct {
	DockerConfigId   *int       `gorm:"column:docker_config_id;primaryKey;autoIncrement"`
	DockerConfigUuid *uuid.UUID `gorm:"column:docker_config_uuid;type:uuid;default:gen_random_uuid()"`

	TemplateId int `gorm:"column:template_id;not null"`

	Registry  *string    `gorm:"type:varchar(255)"`
	ImageName *string    `gorm:"type:varchar(255);not null"`
	CreatedAt *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	Template *ServiceTemplate `gorm:"foreignKey:TemplateId;references:ServiceTemplateId;constraint:OnDelete:CASCADE"`
}

// ===========================
// AdvancedConfig
// ===========================
type AdvancedConfig struct {
	AdvancedConfigId   *int       `gorm:"column:advanced_config_id;primaryKey;autoIncrement"`
	AdvancedConfigUuid *uuid.UUID `gorm:"column:advanced_config_uuid;type:uuid;default:gen_random_uuid()"`

	TemplateId int `gorm:"column:template_id;not null"`

	EnableAuthentication *bool      `gorm:"default:false"`
	GenerateSwaggerDocs  *bool      `gorm:"default:false"`
	CreatedAt            *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt            *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	Template *ServiceTemplate `gorm:"foreignKey:TemplateId;references:ServiceTemplateId;constraint:OnDelete:CASCADE"`
}
