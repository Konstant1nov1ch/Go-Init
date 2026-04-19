package kafka

type Topic struct {
	Id        string `yaml:"id"`
	Name      string `yaml:"name"`
	IsEnabled bool   `yaml:"is_enabled"`
}

type Config struct {
	Enabled bool     `yaml:"enabled" default:"true"`
	Address []string `yaml:"addresses"`

	ConsumerConfig struct {
		Enabled    bool   `yaml:"enabled" default:"false"`
		Topic      Topics `yaml:"topics"`
		GroupId    string `yaml:"group_id"`
		AutoCommit bool   `yaml:"auto_commit" default:"false"`
	} `yaml:"consumer_config"`

	ProducerConfig struct {
		Enabled bool   `yaml:"enabled" default:"false"`
		Topic   Topics `yaml:"topics"`
	} `yaml:"producer_config"`
}

type Topics []Topic

func (t Topics) collectTopics() ([]string, map[string]TopicEvent) {
	var topics []string
	topicInfo := make(map[string]TopicEvent)

	for _, topic := range t {
		topics = append(topics, topic.Name)
		topicInfo[topic.Id] = TopicEvent{
			Name:    topic.Name,
			Enabled: topic.IsEnabled,
		}
	}

	return topics, topicInfo
}
