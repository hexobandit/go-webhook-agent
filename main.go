package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	appIDStr := os.Getenv("APP_ID")
	if appIDStr == "" {
		log.Fatal("APP_ID is required")
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		log.Fatalf("APP_ID must be numeric: %v", err)
	}

	keyPath := os.Getenv("PRIVATE_KEY_PATH")
	if keyPath == "" {
		log.Fatal("PRIVATE_KEY_PATH is required")
	}

	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("WEBHOOK_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app, err := NewGitHubApp(appID, keyPath)
	if err != nil {
		log.Fatalf("Failed to init GitHub App: %v", err)
	}

	http.HandleFunc("/webhook", webhookHandler(app, []byte(secret)))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("Webhook agent listening on :%s (App ID: %d)", port, appID)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
