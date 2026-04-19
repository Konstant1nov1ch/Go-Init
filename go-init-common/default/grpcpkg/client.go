package grpcpkg

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient представляет gRPC-клиент.
type GRPCClient struct {
	conn *grpc.ClientConn
}

// NewGRPCClient создает gRPC-клиент.
func NewGRPCClient(config ClientConfig) (*GRPCClient, error) {
	var creds credentials.TransportCredentials
	if config.UseTLS {
		// Загружаем путь к сертификату из конфига или используем значение по умолчанию
		certPath := config.CACertPath
		if certPath == "" {
			certPath = "./certs/ca.crt" // Дефолтный путь
		}

		// Загрузка корневого сертификата сервера
		caCert, err := os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate from %s: %w", certPath, err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate from %s", certPath)
		}
		tlsConfig := &tls.Config{
			RootCAs: caCertPool,
		}
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	client, err := grpc.NewClient(config.Address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for %s: %w", config.Address, err)
	}

	return &GRPCClient{conn: client}, nil
}

// GetConnection returns the underlying gRPC connection
func (c *GRPCClient) GetConnection() *grpc.ClientConn {
	return c.conn
}
