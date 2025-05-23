package models

import (
	db "gitlab.com/go-init/go-init-common/default/db/pg/orm"
)

// ==================================
// ServiceTemplate methods
// ==================================
func (m *ServiceTemplate) String() string {
	return db.ModelToString(m)
}

func (m *ServiceTemplate) Name() string {
	return "ServiceTemplate"
}

func (m *ServiceTemplate) GenericID() db.GenericID {
	// Возвращаем автоинкрементный integer PK
	return m.ServiceTemplateId
}

// ==================================
// Endpoint methods
// ==================================
func (m *Endpoint) String() string {
	return db.ModelToString(m)
}

func (m *Endpoint) Name() string {
	return "Endpoint"
}

func (m *Endpoint) GenericID() db.GenericID {
	return m.EndpointId
}

// ==================================
// DatabaseConfig methods
// ==================================
func (m *DatabaseConfig) String() string {
	return db.ModelToString(m)
}

func (m *DatabaseConfig) Name() string {
	return "DatabaseConfig"
}

func (m *DatabaseConfig) GenericID() db.GenericID {
	return m.DatabaseConfigId
}

// ==================================
// DockerConfig methods
// ==================================
func (m *DockerConfig) String() string {
	return db.ModelToString(m)
}

func (m *DockerConfig) Name() string {
	return "DockerConfig"
}

func (m *DockerConfig) GenericID() db.GenericID {
	return m.DockerConfigId
}

// ==================================
// AdvancedConfig methods
// ==================================
func (m *AdvancedConfig) String() string {
	return db.ModelToString(m)
}

func (m *AdvancedConfig) Name() string {
	return "AdvancedConfig"
}

func (m *AdvancedConfig) GenericID() db.GenericID {
	return m.AdvancedConfigId
}
