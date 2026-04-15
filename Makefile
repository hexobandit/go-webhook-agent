.PHONY: run build smee gh-forward dev clean

# Load .env if it exists
ifneq (,$(wildcard ./.env))
include .env
export
endif

run:
	go run .

build:
	go build -o bin/webhook-agent .

# Local dev: forward GitHub webhooks via smee.io
# 1. Go to https://smee.io and click "Start a new channel"
# 2. Copy the URL into .env as SMEE_URL
# 3. Use that URL as the webhook URL when configuring the GitHub webhook
smee:
	npx smee-client -u $(SMEE_URL) -t http://localhost:$(PORT)/webhook

# Local dev: forward via GitHub CLI (alternative to smee)
# Requires: gh extension install cli/gh-webhook
# Usage: make gh-forward REPO=owner/repo
#    or: make gh-forward ORG=myorg
gh-forward:
ifndef REPO
ifndef ORG
	$(error Set REPO=owner/repo or ORG=myorg — e.g. make gh-forward REPO=monalisa/smile)
endif
endif
	gh webhook forward \
		--events="pull_request,issues" \
		$(if $(REPO),--repo=$(REPO)) \
		$(if $(ORG),--org=$(ORG)) \
		--secret=$(WEBHOOK_SECRET) \
		--url=http://localhost:$(PORT)/webhook

# Run server + smee proxy together
dev:
	@echo "Starting webhook agent + smee proxy..."
	@make smee &
	@make run

clean:
	rm -rf bin/
