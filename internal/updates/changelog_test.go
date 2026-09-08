package updates

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/rs/zerolog"
	wupdater "github.com/wailsapp/wails/v3/pkg/updater"
)

func coordinatorWithDoer(d HTTPDoer) *Coordinator {
	return New(&fakeEngine{rel: (*wupdater.Release)(nil)}, d, false, zerolog.Nop())
}

func TestChangelog_ReturnsReleaseBody(t *testing.T) {
	c := coordinatorWithDoer(doerFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://api.github.com/repos/"+RepoSlug+"/releases/latest" {
			t.Fatalf("unexpected URL %s", r.URL)
		}
		return jsonResponse(200, `{"body":"## What changed\n- stuff"}`), nil
	}))

	got := c.Changelog(context.Background())
	if got != "## What changed\n- stuff" {
		t.Fatalf("Changelog = %q, want the release body", got)
	}
}

func TestChangelog_StripsEverythingOutsideTheMarkers(t *testing.T) {
	body := "## Welcome\n\nDownload and install instructions here.\n\n" +
		changelogStartMarker + "\n### Highlights\n- did a thing\n" + changelogEndMarker +
		"\n\n<details>install steps</details>"
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	c := coordinatorWithDoer(doerFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, string(payload)), nil
	}))

	got := c.Changelog(context.Background())
	want := "### Highlights\n- did a thing"
	if got != want {
		t.Fatalf("Changelog = %q, want %q", got, want)
	}
}

func TestChangelog_OfflineFallsBack(t *testing.T) {
	c := coordinatorWithDoer(doerFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("dial tcp: no route to host")
	}))

	if got := c.Changelog(context.Background()); got != ChangelogUnavailable {
		t.Fatalf("Changelog = %q, want %q", got, ChangelogUnavailable)
	}
}

func TestChangelog_HTTPErrorFallsBack(t *testing.T) {
	c := coordinatorWithDoer(doerFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(404, `{"message":"Not Found"}`), nil
	}))

	if got := c.Changelog(context.Background()); got != ChangelogUnavailable {
		t.Fatalf("Changelog = %q, want %q", got, ChangelogUnavailable)
	}
}

func TestChangelog_EmptyBodyGivesPlaceholder(t *testing.T) {
	c := coordinatorWithDoer(doerFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"body":""}`), nil
	}))

	got := c.Changelog(context.Background())
	if got == "" || got == ChangelogUnavailable {
		t.Fatalf("Changelog = %q, want a non-empty 'no notes' placeholder", got)
	}
}
