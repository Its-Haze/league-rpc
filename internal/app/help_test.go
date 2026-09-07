package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/logging"
)

func TestApp_Logs_NilRingDegradesGracefully(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	if got := a.GetRecentLogs(); got != nil {
		t.Errorf("GetRecentLogs() = %v, want nil without WithLogs", got)
	}
	if got := a.SubscribeLogs(); got != nil {
		t.Error("SubscribeLogs() returned a channel without WithLogs")
	}
	if err := a.OpenLogsFolder(); err == nil {
		t.Error("OpenLogsFolder should error without a logs directory")
	}
}

func TestApp_Logs_WiredRing(t *testing.T) {
	ring := logging.NewRing(10)
	_, _ = ring.Write([]byte("hello\n"))

	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{}, WithLogs(ring, "C:\\fake\\logs"))

	got := a.GetRecentLogs()
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("GetRecentLogs() = %v, want [\"hello\"]", got)
	}

	sub := a.SubscribeLogs()
	if sub == nil {
		t.Fatal("SubscribeLogs() returned nil with a wired ring")
	}
	_, _ = ring.Write([]byte("world\n"))
	if line := <-sub; line != "world" {
		t.Errorf("subscribed line = %q, want %q", line, "world")
	}
}

func TestApp_OpenLogsFolder_DelegatesToInjectedOpener(t *testing.T) {
	ring := logging.NewRing(10)
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{}, WithLogs(ring, "C:\\fake\\logs"))

	var openedPath string
	a.openFolder = func(path string) error {
		openedPath = path
		return nil
	}
	if err := a.OpenLogsFolder(); err != nil {
		t.Fatalf("OpenLogsFolder: %v", err)
	}
	if openedPath != "C:\\fake\\logs" {
		t.Errorf("openFolder called with %q, want the logs directory", openedPath)
	}

	a.openFolder = func(string) error { return errors.New("boom") }
	if err := a.OpenLogsFolder(); err == nil {
		t.Error("OpenLogsFolder did not propagate the opener's error")
	}
}

func TestApp_GetDiagnostics_ContainsKeyFields(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	got := a.GetDiagnostics()
	for _, want := range []string{"Version:", "OS:", "League process:", "Discord:", "Last update error: none"} {
		if !strings.Contains(got, want) {
			t.Errorf("diagnostics missing %q:\n%s", want, got)
		}
	}
}
