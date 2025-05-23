package eventdata

type TemplateEventData struct {
	Name      string               `json:"name"`
	Endpoints []*EndpointEventData `json:"endpoints"`
	Database  DatabaseEventData    `json:"database"`
	Docker    DockerEventData      `json:"docker"`
	Advanced  *AdvancedEventData   `json:"advanced,omitempty"`
}

type EndpointEventData struct {
	Protocol string            `json:"protocol"`
	Role     string            `json:"role"`
	Config   map[string]string `json:"config,omitempty"`
}

type DatabaseEventData struct {
	Type       string `json:"type"`
	DDL        string `json:"ddl,omitempty"`
	Migrations bool   `json:"migrations,omitempty"`
	Models     bool   `json:"models,omitempty"`
}

type DockerEventData struct {
	Registry  string `json:"registry,omitempty"`
	ImageName string `json:"imageName"`
}

type AdvancedEventData struct {
	EnableAuthentication bool   `json:"enableAuthentication,omitempty"`
	GenerateSwaggerDocs  bool   `json:"generateSwaggerDocs,omitempty"`
	ModulePath           string `json:"modulePath,omitempty"`
	ServiceDescription   string `json:"serviceDescription,omitempty"`
	EnableGraphQL        bool   `json:"enableGraphQL,omitempty"`
	EnableGRPC           bool   `json:"enableGRPC,omitempty"`
}

type ProcessTemplate struct {
	ID     string            `json:"id"`
	Status string            `json:"status"`
	Data   TemplateEventData `json:"data"`
}

const JsonSchema = `
{
	"type": "object",
	"properties": {
		"name": {"type": "string"},
		"endpoints": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"protocol": {"type": "string"},
					"role": {"type": "string"},
					"config": {
						"type": "object",
						"additionalProperties": {"type": "string"}
					}
				}
			}
		},
		"database": {
			"type": "object",
			"properties": {
				"type": {"type": "string"},
				"ddl": {"type": "string"},
				"migrations": {"type": "boolean"},
				"models": {"type": "boolean"}
			}
		},
		"docker": {
			"type": "object",
			"properties": {
				"registry": {"type": "string"},
				"imageName": {"type": "string"}
			}
		},
		"advanced": {
			"type": "object",
			"properties": {
				"enableAuthentication": {"type": "boolean"},
				"generateSwaggerDocs": {"type": "boolean"},
				"modulePath": {"type": "string"},
				"serviceDescription": {"type": "string"},
				"enableGraphQL": {"type": "boolean"},
				"enableGRPC": {"type": "boolean"}
			}
		}
	}
}
`
