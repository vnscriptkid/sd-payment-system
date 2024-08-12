package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// DB setup (replace with your connection details)
const (
	Host     = "localhost"
	Port     = 5432
	User     = "youruser"
	Password = "yourpassword"
	DBName   = "yourdb"
)

func main() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Host, Port, User, Password, DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create idempotency table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS idempotency (
		idempotency_key VARCHAR(255) PRIMARY KEY,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	idempotencyKey := "unique-request-id-123"
	response, err := ProcessRequest(db, idempotencyKey)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(response)
}

// ProcessRequest handles the idempotency check and request processing
func ProcessRequest(db *sql.DB, idempotencyKey string) (string, error) {
	// Attempt to insert the idempotency key
	insertQuery := `
	INSERT INTO idempotency (idempotency_key)
	VALUES ($1)`

	_, err := db.Exec(insertQuery, idempotencyKey)
	if err != nil {
		// Check if the error is a unique constraint violation
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			// 23505 is the error code for unique_violation in PostgreSQL
			return "Request already processed", nil
		}
		return "", err
	}

	// Simulate processing the request
	responseData := "Success: Processed request"

	return responseData, nil
}
