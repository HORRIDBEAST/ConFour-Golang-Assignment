package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var producer *kafka.Producer
var kafkaTopic = "game-events"

// InitKafka initializes the Kafka producer.
func InitKafka() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	var err error
	producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})

	if err != nil {
		log.Printf("Failed to create Kafka producer: %v. Analytics will be disabled.", err)
		producer = nil
		return
	}

	log.Println("Kafka producer initialized.")

	// Go routine for handling delivery reports
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Kafka delivery failed: %v\n", ev.TopicPartition.Error)
				}
			}
		}
	}()
}

// CloseKafka flushes and closes the Kafka producer.
func CloseKafka() {
	if producer != nil {
		producer.Flush(15 * 1000)
		producer.Close()
		log.Println("Kafka producer closed.")
	}
}

// ProduceEvent sends an analytics event to the Kafka topic.
func ProduceEvent(eventType string, data map[string]interface{}) {
	if producer == nil {
		return // Kafka is not initialized
	}

	event := map[string]interface{}{
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal Kafka event: %v", err)
		return
	}

	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kafkaTopic, Partition: int32(kafka.PartitionAny)},
		Value:          payload,
	}, nil)

	if err != nil {
		log.Printf("Failed to produce Kafka message: %v", err)
	}
}
