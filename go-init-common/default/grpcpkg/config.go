package grpcpkg

// ServerConfig содержит параметры настройки gRPC-сервера.
type ServerConfig struct {
	Port string
}

// ClientConfig содержит параметры настройки gRPC-клиента.
type ClientConfig struct {
	ServiceID       string
	Address         string
	NegotiationType string
	UseTLS          bool
	CACertPath      string
}
