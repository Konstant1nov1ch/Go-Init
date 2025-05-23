package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server представляет обертку над gRPC сервером
type Server struct {
	server *grpc.Server
	port   string
}

// ServerConfig содержит конфигурацию сервера
type ServerConfig struct {
	Port string
}

// NewServer создает новый экземпляр сервера
func NewServer(config ServerConfig) (*Server, error) {
	server := grpc.NewServer()

	// Регистрируем сервис здоровья
	RegisterHealthService(server)

	// Регистрируем reflection сервис для удобства отладки
	reflection.Register(server)

	return &Server{
		server: server,
		port:   config.Port,
	}, nil
}

// Start запускает gRPC сервер
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	return s.server.Serve(listener)
}

// Stop останавливает gRPC сервер
func (s *Server) Stop() {
	s.server.GracefulStop()
}

// GetGRPCServer возвращает внутренний gRPC сервер для регистрации сервисов
func (s *Server) GetGRPCServer() *grpc.Server {
	return s.server
}
