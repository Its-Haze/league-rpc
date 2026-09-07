// Package discordapp resolves a Discord Application ID to its public display
package discordapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const lookupTimeout = 5 * time.Second

// HTTPDoer is anything that can execute an *http.Request, satisfied by
// *http.Client. Injected so tests never hit discord.com.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewProductionHTTPDoer returns the client used for the real lookup.
func NewProductionHTTPDoer() HTTPDoer {
	return &http.Client{Timeout: lookupTimeout}
}

// Lookup resolves a Discord Application ID to its public metadata.
type Lookup struct {
	doer HTTPDoer
}

// New builds a Lookup over doer.
func New(doer HTTPDoer) *Lookup {
	return &Lookup{doer: doer}
}

// Name returns the application's display name, or an error if appID does not
// resolve to a real application (bad ID, RPC disabled, or Discord unreachable).
func (l *Lookup) Name(ctx context.Context, appID string) (string, error) {
	url := "https://discord.com/api/v10/applications/" + appID + "/rpc"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := l.doer.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("discord application lookup: HTTP %d: %s", resp.StatusCode, body)
	}

	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode application: %w", err)
	}
	if payload.Name == "" {
		return "", fmt.Errorf("application has no name")
	}
	return payload.Name, nil
}
