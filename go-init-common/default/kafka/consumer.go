package kafka

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/twmb/franz-go/pkg/kgo"
)

type ConsumerConfig struct {
	Event
	workers ConsumerWorkers
}

type ConsumerWorkers map[string][]ConsumerWorker

type ConsumerWorker interface {
	Work(ctx context.Context, value []byte) error
}

const timeOut = 10 * time.Second

func (c *ClientConfig) Start(ctx context.Context) {
	if !c.consumerConfig.enabled {
		return
	}

	for {
		fetches := c.Client.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			c.logger.Error(fmt.Sprintf("Error polling fetches: %v", err))
			continue
		}

		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, record := range p.Records {
				workers := c.getWorkers(record.Topic)
				if len(workers) > 0 {
					if err := c.processRecord(ctx, record, workers); err != nil {
						c.logger.Error(fmt.Sprintf("Error processing record topic=%s, partition=%d, offset=%d: %v", record.Topic, record.Partition, record.Offset, err))
					} else {
						if !c.consumerConfig.autoCommit {
							if err := c.Client.CommitRecords(ctx, record); err != nil {
								c.logger.Error(fmt.Sprintf("Error committing record topic=%s, partition=%d, offset=%d: %v", record.Topic, record.Partition, record.Offset, err))
							}
						}
					}
				} else {
					if !c.consumerConfig.autoCommit {
						if err := c.Client.CommitRecords(ctx, record); err != nil {
							c.logger.Error(fmt.Sprintf("Error committing unhandled record topic=%s, partition=%d, offset=%d: %v", record.Topic, record.Partition, record.Offset, err))
						}
					}
				}
			}
		})
	}
}

func (c *ClientConfig) getWorkers(topic string) []ConsumerWorker {
	return c.consumerConfig.workers[topic]
}

func (c *ClientConfig) processRecord(ctx context.Context, record *kgo.Record, workers []ConsumerWorker) error {
	errGroup, ctx := errgroup.WithContext(ctx)
	for _, worker := range workers {
		worker := worker // capture range variable
		errGroup.Go(func() error {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error(fmt.Sprintf("Recovered from panic: %v", r))
				}
			}()
			return worker.Work(ctx, record.Value)
		})
	}
	if err := errGroup.Wait(); err != nil {
		c.logger.Error(fmt.Sprintf("Error processing record: %v", err))
		return err
	}
	return nil
}

func (c *ClientConfig) RegisterConsumerWorkersByTopic(topicID string, workers ...ConsumerWorker) error {
	if !c.consumerConfig.enabled {
		return fmt.Errorf("consumer is not enabled")
	}
	topic, err := c.consumerConfig.GetTopic(topicID)
	if err != nil {
		return fmt.Errorf("failed to get topic: %w", err)
	}
	c.consumerConfig.workers[topic.Name] = workers
	return nil
}

func (c *ClientConfig) RegisterConsumerWorkers(workers ConsumerWorkers) error {
	if !c.consumerConfig.enabled {
		return fmt.Errorf("consumer is not enabled")
	}
	c.consumerConfig.workers = workers
	return nil
}

func (c *ClientConfig) ConsumeWorkers(topicID string) ([]ConsumerWorker, error) {
	topic, err := c.consumerConfig.GetTopic(topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic: %w", err)
	}
	return c.consumerConfig.workers[topic.Name], nil
}
