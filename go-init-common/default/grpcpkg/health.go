package grpcpkg

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// RegisterHealthService регистрирует HealthCheck сервис.
func RegisterHealthService(server *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
}

// IsHealthy проверяет состояние сервиса через gRPC HealthCheck.
func (c *GRPCClient) IsHealthy(ctx context.Context) (bool, error) {
	healthClient := grpc_health_v1.NewHealthClient(c.conn)
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return false, err
	}
	return resp.Status == grpc_health_v1.HealthCheckResponse_SERVING, nil
}
