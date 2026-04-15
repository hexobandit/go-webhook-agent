package main

import (
	"log"
	"net/http"

	"github.com/google/go-github/v62/github"
)

func webhookHandler(app *GitHubApp, secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := github.ValidatePayload(r, secret)
		if err != nil {
			log.Printf("Invalid payload: %v", err)
			http.Error(w, "invalid payload", http.StatusUnauthorized)
			return
		}

		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			log.Printf("Failed to parse webhook: %v", err)
			http.Error(w, "parse error", http.StatusBadRequest)
			return
		}

		eventType := github.WebHookType(r)
		log.Printf("Received event: %s", eventType)

		// Extract installation ID — all GitHub App webhook payloads include this
		var installationID int64
		switch e := event.(type) {
		case *github.PullRequestEvent:
			installationID = e.GetInstallation().GetID()
		case *github.IssuesEvent:
			installationID = e.GetInstallation().GetID()
		default:
			log.Printf("Ignoring event type: %s", eventType)
			w.WriteHeader(http.StatusAccepted)
			return
		}

		if installationID == 0 {
			log.Printf("No installation ID in %s event", eventType)
			http.Error(w, "missing installation", http.StatusBadRequest)
			return
		}

		client, err := app.ClientForInstallation(r.Context(), installationID)
		if err != nil {
			log.Printf("Failed to get installation client: %v", err)
			http.Error(w, "auth error", http.StatusInternalServerError)
			return
		}

		switch e := event.(type) {
		case *github.PullRequestEvent:
			go handlePullRequest(client, e)
		case *github.IssuesEvent:
			go handleIssue(client, e)
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
