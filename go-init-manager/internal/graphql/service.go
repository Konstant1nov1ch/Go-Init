package graphql

import (
	dbRepo "go-init/internal/database"

	database "gitlab.com/go-init/go-init-common/default/db/pg/orm"

	"gitlab.com/go-init/go-init-common/default/kafka"
	"gitlab.com/go-init/go-init-common/default/logger"
)

type Service struct {
	logger        *logger.Logger
	serviceName   string
	agent         *database.AgentImpl
	dbManagerRepo dbRepo.GoInitManagerRepository
	KafkaProducer *kafka.ClientConfig
}

func New(log *logger.Logger,
	name string,
	dbManagerRepo dbRepo.GoInitManagerRepository,
	agent *database.AgentImpl,
	KafkaProducer *kafka.ClientConfig,
) *Service {
	return &Service{
		logger:        log,
		serviceName:   name,
		dbManagerRepo: dbManagerRepo,
		agent:         agent,
		KafkaProducer: KafkaProducer,
	}
}
