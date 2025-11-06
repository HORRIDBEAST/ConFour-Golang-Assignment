package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Event struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

func main() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          "game-analytics",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	err = consumer.Subscribe("game-events", nil)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	log.Println("Kafka consumer started. Listening for events...")

	// Handle shutdown gracefully
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating\n", sig)
			run = false

		default:
			ev := consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				handleEvent(e.Value)

			case kafka.Error:
				log.Printf("Error: %v\n", e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			}
		}
	}
}

func handleEvent(data []byte) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return
	}

	switch event.Type {
	case "game_started":
		log.Printf("Game Started: %+v", event.Data)
		// Store in database or process analytics

	case "move_made":
		log.Printf("Move Made: %+v", event.Data)
		// Track move patterns

	case "game_ended":
		log.Printf("Game Ended: %+v", event.Data)
		// Update analytics metrics
		processGameEndAnalytics(event.Data)

	default:
		log.Printf("Unknown event type: %s", event.Type)
	}
}

func processGameEndAnalytics(data map[string]interface{}) {
	// Calculate and store analytics
	gameID := data["gameId"]
	winner := data["winner"]
	duration := data["duration"]
	isBot := data["isBot"]

	log.Printf("Analytics - Game: %v, Winner: %v, Duration: %.2fs, IsBot: %v",
		gameID, winner, duration, isBot)

	// Here you would:
	// 1. Store raw event data in database
	// 2. Update aggregated metrics (games per hour, avg duration, etc.)
	// 3. Update player statistics
	// 4. Generate reports
}
