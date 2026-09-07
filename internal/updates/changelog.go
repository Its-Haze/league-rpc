package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChangelogUnavailable is returned in place of release notes when GitHub is
// unreachable, so the GUI has a stable string to render.
const ChangelogUnavailable = "changelog unavailable"

const changelogTimeout = 8 * time.Second

// HTTPDoer is anything that can execute an *http.Request, satisfied by
// *http.Client. Injected so tests never hit api.github.com.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewProductionHTTPDoer returns the client used for the real changelog fetch.
func NewProductionHTTPDoer() HTTPDoer {
	return &http.Client{Timeout: changelogTimeout}
}

// changelogFetcher pulls the latest release body from the GitHub REST API.
type changelogFetcher struct {
	doer HTTPDoer
	repo string
}

func newChangelogFetcher(doer HTTPDoer, repo string) *changelogFetcher {
	return &changelogFetcher{doer: doer, repo: repo}
}

// changelogStartMarker and changelogEndMarker bound the curated
const (
	changelogStartMarker = "<!-- changelog:start -->"
	changelogEndMarker   = "<!-- changelog:end -->"
)

// extractChangelog returns the text between the changelog markers, trimmed.
func extractChangelog(body string) string {
	start := strings.Index(body, changelogStartMarker)
	if start < 0 {
		return body
	}
	start += len(changelogStartMarker)
	end := strings.Index(body[start:], changelogEndMarker)
	if end < 0 {
		return body
	}
	return strings.TrimSpace(body[start : start+end])
}

// fetch returns the latest release's Markdown body, trimmed down to the
func (f *changelogFetcher) fetch(ctx context.Context) (string, error) {
	url := "https://api.github.com/repos/" + f.repo + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := f.doer.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("github releases API: HTTP %d: %s", resp.StatusCode, body)
	}

	var payload struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode release: %w", err)
	}
	if payload.Body == "" {
		return "No release notes for this version.", nil
	}
	return extractChangelog(payload.Body), nil
}
