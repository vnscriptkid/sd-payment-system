package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// Redis client
func getRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
		DB:   0,                // Use default DB
	})
	return rdb
}

// ProcessRequest simulates processing a request with idempotency
func ProcessRequest(idempotencyKey string) (string, error) {
	client := getRedisClient()
	defer client.Close()

	// Define the TTL for the idempotency key
	ttl := 24 * time.Hour

	// Try to set the idempotency key using SETNX
	// Set if key does not exist, if key already holds a value, no operation is performed (atomic, avoid race conditions)
	set, err := client.SetNX(ctx, idempotencyKey, "processing", ttl).Result()
	if err != nil {
		return "", err
	}

	if !set {
		// Key already exists, return the previous response
		previousResponse, err := client.Get(ctx, idempotencyKey).Result()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Request already processed. Previous response: %s", previousResponse), nil
	}

	// Simulate processing the request
	response, err := processRequest()
	if err != nil {
		// Delete the key in case of an error to allow retry
		client.Del(ctx, idempotencyKey)
		return "", err
	}

	// Store the actual response with the same key
	err = client.Set(ctx, idempotencyKey, response, ttl).Err()
	if err != nil {
		return "", err
	}

	return response, nil
}

func processRequest() (string, error) {
	return "success", nil
}

func main() {
	idempotencyKey := "unique-request-id-123"
	response, err := ProcessRequest(idempotencyKey)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(response)
}
