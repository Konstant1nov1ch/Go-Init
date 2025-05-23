package features

import (
	"strings"

	"go-init-gen/internal/eventdata"
)

// FeatureSet represents the detected features in a template
type FeatureSet struct {
	HasGRPC      bool
	HasGraphQL   bool
	HasREST      bool
	HasHTTP      bool // Combination of REST, GraphQL or HTTP
	HasKafka     bool
	HasDatabase  bool
	DatabaseType string
}

// DetectFeatures analyzes the template data and identifies all enabled features
func DetectFeatures(data *eventdata.TemplateEventData) *FeatureSet {
	fs := &FeatureSet{}

	// Check endpoints
	for _, endpoint := range data.Endpoints {
		protocol := normalizeProtocol(endpoint.Protocol)
		role := normalizeRole(endpoint.Role)

		if protocol == ProtocolGRPC && role == RoleServer {
			fs.HasGRPC = true
		}

		if protocol == ProtocolGraphQL && role == RoleServer {
			fs.HasGraphQL = true
			fs.HasHTTP = true
		}

		if protocol == ProtocolREST && role == RoleServer {
			fs.HasREST = true
			fs.HasHTTP = true
		}

		if protocol == ProtocolHTTP && role == RoleServer {
			fs.HasHTTP = true
		}

		if protocol == ProtocolKafka {
			fs.HasKafka = true
		}
	}

	// Check database
	if data.Database.Type != "" && strings.ToLower(data.Database.Type) != "none" {
		fs.HasDatabase = true
		fs.DatabaseType = normalizeDBType(data.Database.Type)
	}

	return fs
}

// normalizeProtocol ensures consistent protocol naming regardless of case
func normalizeProtocol(protocol string) string {
	return strings.ToUpper(protocol)
}

// normalizeRole ensures consistent role naming regardless of case
func normalizeRole(role string) string {
	return strings.ToLower(role)
}

// normalizeDBType ensures consistent database type naming
func normalizeDBType(dbType string) string {
	dbTypeLower := strings.ToLower(dbType)

	// Map some common variations to standard names
	switch dbTypeLower {
	case "postgres", "postgresql":
		return DatabaseTypePostgresql
	case "mysql":
		return DatabaseTypeMysql
	case "mongodb", "mongo":
		return DatabaseTypeMongoDB
	case "redis":
		return DatabaseTypeRedis
	case "none", "":
		return DatabaseTypeNone
	default:
		return dbTypeLower
	}
}

// HasPostgres returns true if the database type is PostgreSQL
func (fs *FeatureSet) HasPostgres() bool {
	return fs.HasDatabase && (fs.DatabaseType == DatabaseTypePostgresql || fs.DatabaseType == DatabaseTypePostgres)
}

// HasMySQL returns true if the database type is MySQL
func (fs *FeatureSet) HasMySQL() bool {
	return fs.HasDatabase && fs.DatabaseType == DatabaseTypeMysql
}

// HasMongoDB returns true if the database type is MongoDB
func (fs *FeatureSet) HasMongoDB() bool {
	return fs.HasDatabase && fs.DatabaseType == DatabaseTypeMongoDB
}

// HasRedis returns true if the database type is Redis
func (fs *FeatureSet) HasRedis() bool {
	return fs.HasDatabase && fs.DatabaseType == DatabaseTypeRedis
}

// HasServerEndpoints returns true if there are any server endpoints
func (fs *FeatureSet) HasServerEndpoints() bool {
	return fs.HasGRPC || fs.HasGraphQL || fs.HasREST || fs.HasHTTP
}
