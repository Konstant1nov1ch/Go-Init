package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// RegisterHealthService регистрирует сервис здоровья
func RegisterHealthService(server *grpc.Server) {
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
}
