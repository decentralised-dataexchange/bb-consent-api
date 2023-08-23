package kafkaUtils

import (
	"github.com/bb-consent/api/src/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type kafkaProducerClient struct {
	Producer *kafka.Producer
}

var KafkaProducerClient kafkaProducerClient

// Init Initialises the kafka producer client
func Init(config *config.Configuration) error {
	// Creating a high level apache kafka producer instance
	// https://github.com/edenhill/librdkafka/tree/master/CONFIGURATION.md
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.Webhooks.KafkaConfig.Broker.URL,
	})
	KafkaProducerClient = kafkaProducerClient{Producer: producer}

	if err != nil {
		return err
	}

	return nil
}
