package eventdata

const (
	// ProcessTemplateEventType тип события для обработки шаблона
	ProcessTemplateEventType = "process-template"

	// ProcessingTopicID ID топика для обработки шаблонов
	ProcessingTopicID = "go-init-processing"

	// JsonSchema схема JSON для событий шаблона
	JsonSchema = `
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
						"role": {"type": "string"}
					}
				}
			},
			"database": {
				"type": "object",
				"properties": {
					"type": {"type": "string"},
					"ddl": {"type": "string"}
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
					"generateSwaggerDocs": {"type": "boolean"}
				}
			}
		}
	}
	`
)

type TemplateEventData struct {
	Name      string               `json:"name"`
	Endpoints []*EndpointEventData `json:"endpoints"`
	Database  DatabaseEventData    `json:"database"`
	Docker    DockerEventData      `json:"docker"`
	Advanced  *AdvancedEventData   `json:"advanced,omitempty"`
}

type EndpointEventData struct {
	Protocol string `json:"protocol"`
	Role     string `json:"role"`
}

type DatabaseEventData struct {
	Type string `json:"type"`
	DDL  string `json:"ddl,omitempty"`
}

type DockerEventData struct {
	Registry  string `json:"registry,omitempty"`
	ImageName string `json:"imageName"`
}

type AdvancedEventData struct {
	EnableAuthentication bool `json:"enableAuthentication,omitempty"`
	GenerateSwaggerDocs  bool `json:"generateSwaggerDocs,omitempty"`
}

type ProcessTemplate struct {
	ID     string            `json:"id"`
	Status string            `json:"status"`
	Data   TemplateEventData `json:"data"`
}
