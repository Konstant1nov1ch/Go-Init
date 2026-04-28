package config

// TracingConfig задаёт экспорт трейсов в OTLP (Grafana Tempo и др.).
// Если otlp_endpoint пустой, проверяется переменная окружения OTEL_EXPORTER_OTLP_ENDPOINT.
type TracingConfig struct {
	Enabled      bool   `yaml:"enabled"`
	OTLPEndpoint string `yaml:"otlp_endpoint"`
}
