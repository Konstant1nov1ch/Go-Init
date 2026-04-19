package grpcpkg

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer представляет gRPC-сервер.
type GRPCServer struct {
	server *grpc.Server
}

// NewGRPCServer создает и возвращает новый gRPC-сервер.
func NewGRPCServer(config ServerConfig) (*GRPCServer, error) {
	grpcServer := grpc.NewServer()
	RegisterHealthService(grpcServer)
	reflection.Register(grpcServer)

	return &GRPCServer{server: grpcServer}, nil
}

// Start запускает gRPC-сервер.
func (s *GRPCServer) Start(port string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	return s.server.Serve(listener)
}

// Stop останавливает сервер.
func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}
