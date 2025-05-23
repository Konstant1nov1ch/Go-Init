package eventdata

import (
	"encoding/json"
)

const (
	// ArchiveReadyEventType тип события для обработки готового архива
	ArchiveReadyEventType = "archive-ready"

	// ArchiveSchema схема для событий архива
	ArchiveSchema = "go-init-archive-schema"
)

// CloudEvent представляет структуру сообщения CloudEvent
type CloudEvent struct {
	SpecVersion     string          `json:"specversion"`
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	DataContentType string          `json:"datacontenttype"`
	DataSchema      string          `json:"dataschema"`
	Time            string          `json:"time"`
	Data            json.RawMessage `json:"data"`
}

// ArchiveMetadata представляет метаданные архива, полученные от издателя
type ArchiveMetadata struct {
	// ID уникальный идентификатор, используемый для связи этого архива с запросом
	ID string `json:"id,omitempty"`

	// BucketName имя бакета, в котором хранится объект
	BucketName string `json:"bucketName"`

	// ObjectName имя объекта в бакете
	ObjectName string `json:"objectName"`

	// Size размер объекта в байтах
	Size int64 `json:"objectSize"`

	// ContentType MIME-тип объекта
	ContentType string `json:"contentType"`

	// ETag тег сущности объекта
	ETag string `json:"etag"`

	// CreatedAt временная метка создания объекта
	CreatedAt string `json:"createdAt"`

	// LastModified временная метка последнего изменения объекта
	LastModified string `json:"lastModified"`

	// AdditionalMetadata дополнительные метаданные для объекта
	AdditionalMetadata map[string]string `json:"additionalMetadata,omitempty"`

	// ArchiveType указывает тип архива (например, "zip", "tar" и т.д.)
	ArchiveType string `json:"archiveType"`

	// PresignedURL опциональное поле для хранения предварительно подписанного URL для доступа к архиву
	PresignedURL string `json:"presignedURL,omitempty"`

	// ExpiresAt опциональное поле для хранения времени истечения срока действия предварительно подписанного URL
	ExpiresAt string `json:"expiresAt,omitempty"`
}
