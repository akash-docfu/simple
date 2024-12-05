package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func validateAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Try getting from query parameter
			apiKey = r.URL.Query().Get("api_key")
		}

		// Get expected API key from environment
		expectedAPIKey := os.Getenv("API_KEY")

		// Validate API key
		if apiKey == "" || !strings.EqualFold(apiKey, expectedAPIKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{
				Message: "Invalid or missing API key",
				Status:  http.StatusUnauthorized,
			})
			return
		}

		// If API key is valid, proceed to the next handler
		next(w, r)
	}
}

func protected(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Message: "Welcome to the protected endpoint!",
		Status:  http.StatusOK,
	})
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Check if API_KEY is set
	if os.Getenv("API_KEY") == "" {
		log.Fatal("API_KEY must be set in .env file")
	}

	// Define routes
	http.HandleFunc("/api/protected", validateAPIKey(protected))

	// Start server
	port := os.Getenv("PORT_NUM")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
