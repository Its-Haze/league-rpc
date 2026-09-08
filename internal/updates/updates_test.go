package updates

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	wupdater "github.com/wailsapp/wails/v3/pkg/updater"
)

// fakeEngine is a scripted stand-in for *wupdater.Updater.
type fakeEngine struct {
	rel        *wupdater.Release
	checkErr   error
	checkCalls int
	dlCalls    int
	dlErr      error
	restart    int
}

func (f *fakeEngine) Check(context.Context) (*wupdater.Release, error) {
	f.checkCalls++
	return f.rel, f.checkErr
}
func (f *fakeEngine) DownloadAndInstall(context.Context) error { f.dlCalls++; return f.dlErr }
func (f *fakeEngine) Restart(context.Context) error            { f.restart++; return nil }

func newTestCoordinator(eng engine) *Coordinator {
	return New(eng, doerFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("no network in test")
	}), false, zerolog.Nop())
}

func TestCheck_NoUpdate(t *testing.T) {
	c := newTestCoordinator(&fakeEngine{rel: nil})

	var got Status
	c.OnChange(func(s Status) { got = s })

	s, err := c.Check(context.Background())
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if s.Available {
		t.Fatalf("Available = true for a nil release, want false")
	}
	if got.Available {
		t.Fatal("onChange saw Available = true, want false")
	}
}

func TestCheck_UpdateAvailable(t *testing.T) {
	eng := &fakeEngine{rel: &wupdater.Release{Version: "2.0.0", Notes: "big changes"}}
	c := newTestCoordinator(eng)

	var seen Status
	c.OnChange(func(s Status) { seen = s })

	s, err := c.Check(context.Background())
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !s.Available || s.Version != "2.0.0" || s.Notes != "big changes" {
		t.Fatalf("status = %+v, want available 2.0.0 with notes", s)
	}
	if seen != s {
		t.Fatalf("onChange status %+v != returned %+v", seen, s)
	}
	if c.Status() != s {
		t.Fatalf("Status() %+v != returned %+v", c.Status(), s)
	}
}

func TestCheck_ErrorKeepsPriorAvailability(t *testing.T) {
	eng := &fakeEngine{rel: &wupdater.Release{Version: "2.0.0"}}
	c := newTestCoordinator(eng)

	if _, err := c.Check(context.Background()); err != nil {
		t.Fatalf("first Check: %v", err)
	}

	eng.rel = nil
	eng.checkErr = errors.New("network down")
	s, err := c.Check(context.Background())
	if err == nil {
		t.Fatal("expected the check error to propagate")
	}
	if !s.Available || s.Version != "2.0.0" {
		t.Fatalf("status = %+v, want the prior available release retained", s)
	}
	if s.LastError == "" {
		t.Fatal("LastError not recorded")
	}
}

func TestDownload_NoPendingRelease(t *testing.T) {
	c := newTestCoordinator(&fakeEngine{rel: nil})
	err := c.Download(context.Background())
	if err == nil || !strings.Contains(err.Error(), "no update available") {
		t.Fatalf("Download err = %v, want 'no update available'", err)
	}
}

func TestDownload_HappyPathDoesNotRestart(t *testing.T) {
	eng := &fakeEngine{rel: &wupdater.Release{Version: "2.0.0"}}
	c := newTestCoordinator(eng)

	if err := c.Download(context.Background()); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if eng.dlCalls != 1 {
		t.Fatalf("DownloadAndInstall called %d times, want 1", eng.dlCalls)
	}
	if eng.restart != 0 {
		t.Fatal("Download must not restart; that is a separate confirmed step")
	}

	if err := c.Restart(context.Background()); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if eng.restart != 1 {
		t.Fatalf("Restart called %d times, want 1", eng.restart)
	}
}

func TestDevBuild_ChecksAndActionsDisabled(t *testing.T) {
	eng := &fakeEngine{rel: &wupdater.Release{Version: "9.9.9"}}
	c := New(eng, nil, true, zerolog.Nop())

	// Run returns at once and never touches the engine.
	done := make(chan struct{})
	go func() { c.Run(context.Background()); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not return promptly for a dev build")
	}

	s, err := c.Check(context.Background())
	if err != nil || s.Available {
		t.Fatalf("dev Check = (%+v, %v), want empty status and no error", s, err)
	}
	if err := c.Download(context.Background()); err == nil {
		t.Fatal("dev Download should refuse")
	}
	if eng.checkCalls != 0 || eng.dlCalls != 0 {
		t.Fatal("dev build must not call the engine at all")
	}
}

func TestRun_LaunchCheckThenStops(t *testing.T) {
	eng := &fakeEngine{rel: nil}
	c := newTestCoordinator(eng)
	c.interval = time.Hour // keep the ticker out of the way

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { c.Run(ctx); close(done) }()

	// The launch check should land quickly.
	deadline := time.After(2 * time.Second)
	for eng.checkCalls == 0 {
		select {
		case <-deadline:
			t.Fatal("no launch check within 2s")
		case <-time.After(5 * time.Millisecond):
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop on context cancel")
	}
}

// waitFor polls cond every 5ms until it's true or the 2s deadline passes,
// for asserting on state a background goroutine (autoDownload) will reach.
func waitFor(t *testing.T, cond func() bool, msg string) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for !cond() {
		select {
		case <-deadline:
			t.Fatal(msg)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestCheck_TriggersAutoDownloadInTheBackground(t *testing.T) {
	eng := &fakeEngine{rel: &wupdater.Release{Version: "2.0.0"}}
	c := newTestCoordinator(eng)

	if _, err := c.Check(context.Background()); err != nil {
		t.Fatalf("Check: %v", err)
	}

	waitFor(t, func() bool { return eng.dlCalls == 1 }, "auto-download never ran")

	s := c.Status()
	if !s.Ready {
		t.Fatalf("status = %+v, want Ready after a successful auto-download", s)
	}
}

func TestAutoDownload_SkipsOnceAlreadyReady(t *testing.T) {
	eng := &fakeEngine{rel: &wupdater.Release{Version: "2.0.0"}}
	c := newTestCoordinator(eng)

	if _, err := c.Check(context.Background()); err != nil {
		t.Fatalf("first Check: %v", err)
	}
	waitFor(t, func() bool { return eng.dlCalls == 1 }, "auto-download never ran")
	waitFor(t, func() bool { return c.Status().Ready }, "never reached Ready")

	// A later check (e.g. the next periodic tick) must not download again.
	if _, err := c.Check(context.Background()); err != nil {
		t.Fatalf("second Check: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	if eng.dlCalls != 1 {
		t.Fatalf("DownloadAndInstall called %d times, want 1 once already Ready", eng.dlCalls)
	}
}

func TestAutoDownload_FirstFailureStaysSilent(t *testing.T) {
	eng := &fakeEngine{rel: &wupdater.Release{Version: "2.0.0"}, dlErr: errors.New("network blip")}
	c := newTestCoordinator(eng)

	if _, err := c.Check(context.Background()); err != nil {
		t.Fatalf("Check: %v", err)
	}
	waitFor(t, func() bool { return eng.dlCalls == 1 }, "auto-download never ran")
	waitFor(t, func() bool { return !c.Status().Downloading }, "download never settled")

	if s := c.Status(); s.LastError != "" {
		t.Fatalf("LastError = %q after one failure, want empty (stays silent)", s.LastError)
	}
}

func TestAutoDownload_SecondConsecutiveFailureSurfaces(t *testing.T) {
	eng := &fakeEngine{rel: &wupdater.Release{Version: "2.0.0"}, dlErr: errors.New("network blip")}
	c := newTestCoordinator(eng)

	if _, err := c.Check(context.Background()); err != nil {
		t.Fatalf("first Check: %v", err)
	}
	waitFor(t, func() bool { return eng.dlCalls == 1 }, "first auto-download never ran")
	waitFor(t, func() bool { return !c.Status().Downloading }, "first download never settled")

	if _, err := c.Check(context.Background()); err != nil {
		t.Fatalf("second Check: %v", err)
	}
	waitFor(t, func() bool { return eng.dlCalls == 2 }, "second auto-download never ran")

	waitFor(t, func() bool { return c.Status().LastError != "" }, "LastError never surfaced after two failures")
}

// doerFunc adapts a function to HTTPDoer.
type doerFunc func(*http.Request) (*http.Response, error)

func (d doerFunc) Do(req *http.Request) (*http.Response, error) { return d(req) }

// jsonResponse is a helper for changelog tests.
func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
