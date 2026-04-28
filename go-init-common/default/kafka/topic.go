package kafka

import "fmt"

// entity
type Event struct {
	topics     map[string]TopicEvent
	enabled    bool
	autoCommit bool
}

// TopicEntity
type TopicEvent struct {
	Name    string
	Enabled bool
}

func (e *Event) GetTopic(topicID string) (TopicEvent, error) {
	topic, ok := e.topics[topicID]
	if !ok {
		return TopicEvent{}, fmt.Errorf("unknow topic: %s", topic.Name)
	}
	return topic, nil
}
