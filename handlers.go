package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v62/github"
)

func handlePullRequest(client *github.Client, event *github.PullRequestEvent) {
	action := event.GetAction()
	if action != "opened" && action != "synchronize" {
		log.Printf("Skipping PR action: %s", action)
		return
	}

	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	number := event.GetNumber()

	log.Printf("Processing PR %s/%s#%d (action: %s)", owner, repo, number, action)

	// TODO: add your actual checks here (linting, file validation, policy checks, etc.)

	body := fmt.Sprintf(
		"**Webhook Agent** received this PR.\n\n"+
			"| Field | Value |\n"+
			"|-------|-------|\n"+
			"| Action | `%s` |\n"+
			"| Branch | `%s` -> `%s` |\n"+
			"| Author | @%s |\n\n"+
			"Running checks...",
		action,
		event.GetPullRequest().GetHead().GetRef(),
		event.GetPullRequest().GetBase().GetRef(),
		event.GetPullRequest().GetUser().GetLogin(),
	)

	comment := &github.IssueComment{Body: &body}
	_, _, err := client.Issues.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		log.Printf("Failed to comment on PR %s/%s#%d: %v", owner, repo, number, err)
		return
	}
	log.Printf("Commented on PR %s/%s#%d", owner, repo, number)
}

func handleIssue(client *github.Client, event *github.IssuesEvent) {
	if event.GetAction() != "labeled" {
		return
	}

	ctx := context.Background()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	number := event.GetIssue().GetNumber()
	label := event.GetLabel().GetName()

	log.Printf("Processing issue %s/%s#%d (label added: %s)", owner, repo, number, label)

	// TODO: add logic per label (e.g. triage, assign, auto-close, etc.)

	body := fmt.Sprintf(
		"**Webhook Agent** noticed label `%s` was added.\n\nProcessing...",
		label,
	)

	comment := &github.IssueComment{Body: &body}
	_, _, err := client.Issues.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		log.Printf("Failed to comment on issue %s/%s#%d: %v", owner, repo, number, err)
		return
	}
	log.Printf("Commented on issue %s/%s#%d", owner, repo, number)
}
