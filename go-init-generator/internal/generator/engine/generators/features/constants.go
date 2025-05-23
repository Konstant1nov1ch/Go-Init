package features

// Protocol constants
const (
	ProtocolGRPC    = "GRPC"
	ProtocolREST    = "REST"
	ProtocolGraphQL = "GRAPHQL"
	ProtocolKafka   = "KAFKA"
	ProtocolHTTP    = "HTTP"
)

// Role constants
const (
	RoleServer = "server"
	RoleClient = "client"
)

// Database type constants
const (
	DatabaseTypePostgresql = "postgresql"
	DatabaseTypePostgres   = "postgres"
	DatabaseTypeMysql      = "mysql"
	DatabaseTypeMongoDB    = "mongodb"
	DatabaseTypeRedis      = "redis"
	DatabaseTypeNone       = "none"
)

// Feature name constants - used for naming methods or referencing features
const (
	FeatureDatabase = "database"
	FeatureGRPC     = "grpc"
	FeatureGraphQL  = "graphql"
	FeatureHTTP     = "http"
	FeatureREST     = "rest"
	FeatureKafka    = "kafka"
)
