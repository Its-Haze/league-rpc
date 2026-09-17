package lcu

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// TestGopherLoggerWritesDiscoveryFailures covers the case a stalled connection
// left unexplained: lcu-gopher's reason has to reach the app's own log.
func TestGopherLoggerWritesDiscoveryFailures(t *testing.T) {
	var buf bytes.Buffer
	g := gopherLogger{logger: zerolog.New(&buf).Level(zerolog.DebugLevel)}

	g.Debug("connection", "%s discovery failed: %v", "lockfile", "no valid lockfile found")
	out := buf.String()

	if !strings.Contains(out, "lockfile discovery failed: no valid lockfile found") {
		t.Errorf("formatted message missing from log: %s", out)
	}
	if !strings.Contains(out, `"lcu":"connection"`) {
		t.Errorf("endpoint field missing from log: %s", out)
	}
}

// TestGopherLoggerLevels keeps lcu-gopher's errors visible without debug mode.
func TestGopherLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	g := gopherLogger{logger: zerolog.New(&buf).Level(zerolog.InfoLevel)}

	g.Debug("connection", "routine detail")
	g.Error("connection", "something broke")

	out := buf.String()
	if strings.Contains(out, "routine detail") {
		t.Errorf("debug output must be filtered at info level: %s", out)
	}
	if !strings.Contains(out, `"level":"warn"`) || !strings.Contains(out, "something broke") {
		t.Errorf("errors must survive as warnings: %s", out)
	}
}

// TestFormatGopherLeavesBareMessages covers a message whose own text contains
// a percent verb, which fmt would otherwise mangle.
func TestFormatGopherLeavesBareMessages(t *testing.T) {
	if got := formatGopher("100%% done"); got != "100%% done" {
		t.Errorf("formatGopher() = %q, want the message untouched", got)
	}
}
