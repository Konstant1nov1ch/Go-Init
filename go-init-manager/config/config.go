package config

import (
	"gitlab.com/go-init/go-init-common/default/grpcpkg"
	"gitlab.com/go-init/go-init-common/default/kafka"
	"gitlab.com/go-init/go-init-common/default/logger"

	db "gitlab.com/go-init/go-init-common/default/db/pg"
	myserver "gitlab.com/go-init/go-init-common/default/http/server"

	c "gitlab.com/go-init/go-init-common/default/config"

	"github.com/mcuadros/go-defaults"
)

type AppConfig struct {
	Database db.Config            `yaml:"postgres_db"`
	Logger   logger.Config        `yaml:"logger"`
	HttpServ myserver.Config      `yaml:"http_server"`
	GrpcServ grpcpkg.ServerConfig `yaml:"grpc_server"`
	Kafka    kafka.Config         `yaml:"kafka"`
}

func GetConfig() *AppConfig {
	config := &AppConfig{}
	c.OpenConfig(&config)
	defaults.SetDefaults(&config.Logger)
	return config
}
