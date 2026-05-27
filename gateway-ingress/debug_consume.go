package main

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

func main() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:19092"},
		Topic:     "ondc_events_sanitized",
		Partition: 0,
		MaxBytes:  10e6, // 10MB
	})

	r.SetOffset(kafka.LastOffset)

	log.Println("Waiting for new messages...")
	m, err := r.ReadMessage(context.Background())
	if err != nil {
		log.Fatal("failed to read message:", err)
	}
	fmt.Printf("Message: %s\n", string(m.Value))
	r.Close()
}
