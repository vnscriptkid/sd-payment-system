package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const maxRetries = 3

// PaymentMessage represents the structure of a payment message.
type PaymentMessage struct {
	PaymentID  string `json:"payment_id"`
	RetryCount int    `json:"retry_count"`
}

func consumePaymentEvent(consumer *kafka.Consumer, producer *kafka.Producer) {
	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			processPayment(msg.Value, producer)
		} else {
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func processPayment(message []byte, producer *kafka.Producer) {
	var paymentMessage PaymentMessage

	// Decode JSON message into the struct
	err := json.Unmarshal(message, &paymentMessage)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s", err)
		return
	}

	fmt.Printf("Processing payment: %v\n", paymentMessage)

	// Simulate a payment failure
	if isRetryableError() {
		if paymentMessage.RetryCount < maxRetries {
			fmt.Println("Retryable error occurred, sending to retry queue")
			sendToRetryQueue(paymentMessage, producer)
		} else {
			fmt.Println("Max retries reached, sending to dead letter queue")
			sendToDeadLetterQueue(paymentMessage, producer)
		}
	} else {
		fmt.Println("Non-retryable error, handling accordingly")
	}
}

func sendToRetryQueue(paymentMessage PaymentMessage, producer *kafka.Producer) {
	paymentMessage.RetryCount++

	// Encode the struct into JSON
	message, err := json.Marshal(paymentMessage)
	if err != nil {
		log.Fatalf("Failed to marshal message: %s", err)
	}

	topic := "payments_retry"
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}, nil)

	if err != nil {
		log.Fatalf("Failed to produce message to retry queue: %s", err)
	}

	producer.Flush(15 * 1000)
}

func sendToDeadLetterQueue(paymentMessage PaymentMessage, producer *kafka.Producer) {
	// Encode the struct into JSON
	message, err := json.Marshal(paymentMessage)
	if err != nil {
		log.Fatalf("Failed to marshal message: %s", err)
	}

	topic := "payments_dead_letter"
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}, nil)

	if err != nil {
		log.Fatalf("Failed to produce message to dead letter queue: %s", err)
	}

	producer.Flush(15 * 1000)
}

func isRetryableError() bool {
	// Simulate a retryable error
	return time.Now().Unix()%2 == 0
}

func main() {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "payment_group",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}

	defer c.Close()

	c.SubscribeTopics([]string{"payments", "payments_retry"}, nil)

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}

	defer p.Close()

	consumePaymentEvent(c, p)
}
