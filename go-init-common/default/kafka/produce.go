package kafka

import (
	"context"
	"fmt"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type ProducerConfig struct {
	Event
}

type ProduceEvent struct {
	Type          string
	Schema        string
	Source        string
	TopicID       string
	CorrelationID string
	Data          any
}

func (pc *ProduceEvent) SetCorrelationID(correlationID string) {
	pc.CorrelationID = correlationID
}

func (pc *ProduceEvent) SetData(data any) {
	pc.Data = data
}

func (c *ClientConfig) Produce(ctx context.Context, ev *ProduceEvent) {
	if !c.producerConfig.enabled {
		return
	}

	topic, err := c.producerConfig.GetTopic(ev.TopicID)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Error getting topic: %v", err))
		return
	}
	if !topic.Enabled {
		c.logger.Error(fmt.Sprintf("Topic is not enabled: %v", topic))
		return
	}

	event := ce.NewEvent()
	event.SetType(ev.Type)
	event.SetExtension("correlation-id", ev.CorrelationID)
	event.SetSource(ev.Source)
	event.SetDataSchema(ev.Schema)
	event.SetTime(time.Now())
	event.SetID(uuid.New().String())

	err = event.SetData(ce.ApplicationJSON, ev.Data)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Error setting data: %v", err))
		return
	}

	message, err := event.MarshalJSON()
	if err != nil {
		c.logger.Error(fmt.Sprintf("Error marshalling event: %v", err))
		return
	}
	c.produce(topic.Name, message, nil)
}

func (c *ClientConfig) produce(topic string, message []byte, headers map[string][]byte) {
	ctx := context.Background()

	recordHeader := make([]kgo.RecordHeader, 0, len(headers))
	for k, v := range headers {
		recordHeader = append(recordHeader, kgo.RecordHeader{
			Key:   k,
			Value: v,
		})
	}

	records := &kgo.Record{
		Topic:   topic,
		Value:   message,
		Headers: recordHeader,
	}

	c.Client.Produce(ctx, records, func(_ *kgo.Record, err error) {
		if err != nil {
			c.logger.Error(fmt.Sprintf("Error producing event: %v", err))
		}
	})
}

func (c *ClientConfig) ProducerIsEnabled() bool {
	return c.producerConfig.enabled
}

func (c *ClientConfig) ConsumerIsEnabled() bool {
	return c.consumerConfig.enabled
}

func (c *ClientConfig) IsEnabled() bool {
	return c.enabled
}
