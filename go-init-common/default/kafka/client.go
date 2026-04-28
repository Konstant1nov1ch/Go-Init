package kafka

import (
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"gitlab.com/go-init/go-init-common/default/logger"
)

type ClientConfig struct {
	Client         *kgo.Client
	logger         *logger.Logger
	enabled        bool
	consumerConfig *ConsumerConfig
	producerConfig *ProducerConfig
}

func NewClientConfig(config *Config, logger *logger.Logger) (*ClientConfig, error) {
	logger.Info(fmt.Sprintf("kafka config: %+v", config))
	if !config.Enabled {
		return nil, nil
	}

	options := []kgo.Opt{}
	consumer := ConsumerConfig{}
	producer := ProducerConfig{}

	if config.ConsumerConfig.Enabled {
		consumer.enabled = true
		consumer.autoCommit = config.ConsumerConfig.AutoCommit
		var names []string
		names, consumer.topics = config.ConsumerConfig.Topic.collectTopics()
		options = append(options, kgo.ConsumerGroup(config.ConsumerConfig.GroupId),
			kgo.ConsumeTopics(names...),
			kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		)
		consumer.workers = ConsumerWorkers{}
	}

	if config.ProducerConfig.Enabled {
		producer.enabled = true
		_, producer.topics = config.ProducerConfig.Topic.collectTopics()
	}

	if len(config.Address) > 0 {
		options = append(options, kgo.SeedBrokers(config.Address...))
	}

	cl, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	logger.Info("kafka client created")

	return &ClientConfig{
		Client:         cl,
		logger:         logger,
		enabled:        config.Enabled,
		consumerConfig: &consumer,
		producerConfig: &producer,
	}, nil
}
