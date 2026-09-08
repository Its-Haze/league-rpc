package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_FansOutToFileAndRing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)

	sink, err := New(Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer sink.Close()

	sink.Logger.Info().Str("k", "v").Msg("hello world")

	lines := sink.Ring.RecentLines()
	if len(lines) != 1 || !strings.Contains(lines[0], "hello world") {
		t.Fatalf("ring missing the line: %v", lines)
	}

	logPath := filepath.Join(dir, "league-rpc", "logs", LogFileName)
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Fatalf("log file missing the line: %s", data)
	}
}

func TestNew_LevelFollowsDebugFlag(t *testing.T) {
	cases := []struct {
		name       string
		debug      bool
		wantDebugs int
	}{
		{"info level drops debug", false, 0},
		{"debug level keeps debug", true, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("APPDATA", t.TempDir())
			sink, err := New(Options{Debug: tc.debug})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			defer sink.Close()

			sink.Logger.Debug().Msg("a debug line")

			got := 0
			for _, l := range sink.Ring.RecentLines() {
				if strings.Contains(l, "a debug line") {
					got++
				}
			}
			if got != tc.wantDebugs {
				t.Fatalf("debug lines in ring = %d, want %d", got, tc.wantDebugs)
			}
		})
	}
}

func TestSetDebug_ChangesLevelLive(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	sink, err := New(Options{Debug: false})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer sink.Close()

	sink.Logger.Debug().Msg("dropped before SetDebug")
	SetDebug(true)
	sink.Logger.Debug().Msg("kept after SetDebug")

	lines := sink.Ring.RecentLines()
	if containsLine(lines, "dropped before SetDebug") {
		t.Error("a debug line logged before SetDebug(true) should have been dropped")
	}
	if !containsLine(lines, "kept after SetDebug") {
		t.Error("a debug line logged after SetDebug(true) should have been kept")
	}
}

func containsLine(lines []string, substr string) bool {
	for _, l := range lines {
		if strings.Contains(l, substr) {
			return true
		}
	}
	return false
}

func TestLogDir_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)

	got, err := LogDir()
	if err != nil {
		t.Fatalf("LogDir: %v", err)
	}
	if _, err := os.Stat(got); err != nil {
		t.Fatalf("logs dir not created: %v", err)
	}
}
