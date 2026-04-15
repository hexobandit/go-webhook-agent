package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v62/github"
)

type GitHubApp struct {
	AppID      int64
	PrivateKey *rsa.PrivateKey

	mu     sync.Mutex
	tokens map[int64]*cachedToken
}

type cachedToken struct {
	token     string
	expiresAt time.Time
}

func NewGitHubApp(appID int64, keyPath string) (*GitHubApp, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", keyPath)
	}

	var key *rsa.PrivateKey
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		key = k
	} else if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		var ok bool
		key, ok = k.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key is not RSA")
		}
	} else {
		return nil, fmt.Errorf("failed to parse private key")
	}

	return &GitHubApp{
		AppID:      appID,
		PrivateKey: key,
		tokens:     make(map[int64]*cachedToken),
	}, nil
}

func (a *GitHubApp) generateJWT() (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now.Add(-60 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    fmt.Sprintf("%d", a.AppID),
	})
	return token.SignedString(a.PrivateKey)
}

// ClientForInstallation returns a github client authenticated as the given installation.
// Tokens are cached and refreshed automatically.
func (a *GitHubApp) ClientForInstallation(ctx context.Context, installationID int64) (*github.Client, error) {
	a.mu.Lock()
	if c, ok := a.tokens[installationID]; ok && time.Now().Before(c.expiresAt) {
		a.mu.Unlock()
		return github.NewClient(nil).WithAuthToken(c.token), nil
	}
	a.mu.Unlock()

	jwtStr, err := a.generateJWT()
	if err != nil {
		return nil, fmt.Errorf("generating JWT: %w", err)
	}

	jwtClient := github.NewClient(nil).WithAuthToken(jwtStr)
	tok, _, err := jwtClient.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return nil, fmt.Errorf("creating installation token: %w", err)
	}

	a.mu.Lock()
	a.tokens[installationID] = &cachedToken{
		token:     tok.GetToken(),
		expiresAt: tok.GetExpiresAt().Time.Add(-5 * time.Minute),
	}
	a.mu.Unlock()

	return github.NewClient(nil).WithAuthToken(tok.GetToken()), nil
}
