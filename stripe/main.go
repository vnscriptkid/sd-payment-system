package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
)

var db *sql.DB

func init() {
	var err error
	connStr := "user=youruser dbname=yourdb sslmode=disable password=yourpass"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	stripe.Key = "sk_xxx"

	r := mux.NewRouter()

	// Endpoint for checkout
	r.HandleFunc("/checkout", createPayment).Methods("POST")
	// Endpoint for handling webhook
	r.HandleFunc("/webhook", handleWebhook).Methods("POST")
	// Serve completion page explicitly
	r.HandleFunc("/completion", serveCompletionPage).Methods("GET")

	// Serve static files
	staticDir := "./static/"
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(staticDir))))

	http.Handle("/", r)
	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Serve the completion page
func serveCompletionPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "completion.html"))
}

func createPayment(w http.ResponseWriter, r *http.Request) {
	// Assume the order ID and amount come from the request
	amount := int64(2000) // 20.00 USD
	currency := "usd"
	orderID := "order_123"

	// Create a payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Params: stripe.Params{
			Metadata: map[string]string{
				"order_id": orderID,
			},
		},
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store the token in the database
	_, err = db.Exec("INSERT INTO payments (order_id, stripe_payment_intent_id) VALUES ($1, $2)", orderID, intent.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the client secret to be used on the frontend
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(intent.ClientSecret))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement webhook handler
	w.WriteHeader(http.StatusOK)
}
