# go-webhook-agent

GitHub webhook agent that listens for PR and issue events across repos, runs checks, and comments back.

## How it works

```
Other repos (webhook) ──> smee.io / gh forward ──> localhost:3000/webhook
                                                        │
                                              ┌─────────┴─────────┐
                                              │                   │
                                        PR opened           Issue labeled
                                              │                   │
                                        Comment PR         Comment issue
```

NOTE: ```gh webhook forward``` requires a GitHub Teams/Enterprise plan. 


**Events handled:**
- `pull_request` (opened, synchronize) — comments on the PR
- `issues` (labeled) — comments on the issue

## Setup

### 1. Install Go

```bash
brew install go
```

### 2. Configure

```bash
cp .env.example .env
# Edit .env with your values
```

You need:
- **GITHUB_TOKEN** — [PAT](https://github.com/settings/tokens) with `repo` scope
- **WEBHOOK_SECRET** — any random string (e.g. `openssl rand -hex 20`)

### 3. Local dev with smee.io

```bash
# Get a channel URL from https://smee.io
# Add it to .env as SMEE_URL

# Terminal 1: start the agent
make run

# Terminal 2: forward webhooks
make smee
```

### 4. Local dev with gh CLI (alternative)

```bash
gh extension install cli/gh-webhook
make gh-forward
```

### 5. Configure webhooks on other repos

Go to **Settings > Webhooks** on any repo you want to monitor:

| Field | Value |
|-------|-------|
| Payload URL | Your smee.io URL (dev) or public URL (prod) |
| Content type | `application/json` |
| Secret | Same as `WEBHOOK_SECRET` in your `.env` |
| Events | Select: **Pull requests**, **Issues** |

### 6. Deploy (optional)

```bash
docker build -t webhook-agent .
docker run -e GITHUB_TOKEN=... -e WEBHOOK_SECRET=... -p 3000:3000 webhook-agent
```

## Project structure

```
main.go       — entry point, HTTP server
github.go     — GitHub client setup
webhook.go    — payload validation + event routing
handlers.go   — PR and issue event handlers (add your logic here)
```
