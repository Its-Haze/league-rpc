package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// withMigrations swaps the step table for the duration of a test.
func withMigrations(t *testing.T, steps map[int]migrationStep) {
	t.Helper()
	original := migrations
	migrations = steps
	t.Cleanup(func() { migrations = original })
}

func writeConfigFile(t *testing.T, body string) string {
	t.Helper()
	t.Setenv("APPDATA", t.TempDir())
	path, _ := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// Every version below the current one needs a step, or a file written by an
// older build stops mid-walk and loads with the wrong shape.
func TestMigrations_CoverEveryVersionBelowCurrent(t *testing.T) {
	for version := 1; version < CurrentSchemaVersion; version++ {
		if _, ok := migrations[version]; !ok {
			t.Errorf("no migration registered for schema v%d -> v%d", version, version+1)
		}
	}
}

func TestLoad_RunsRegisteredStepsAndWritesTheUpgradeBack(t *testing.T) {
	withMigrations(t, map[int]migrationStep{
		0: func(raw []byte) ([]byte, error) {
			var c Config
			if err := json.Unmarshal(raw, &c); err != nil {
				return nil, err
			}
			c.DiscordAppID = "upgraded"
			return json.Marshal(c)
		},
	})

	path := writeConfigFile(t, `{"discord_app_id": "original", "theme": "dark"}`)

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.DiscordAppID != "upgraded" {
		t.Errorf("DiscordAppID = %q, want the migrated value", got.DiscordAppID)
	}
	if got.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", got.SchemaVersion, CurrentSchemaVersion)
	}

	raw, _ := os.ReadFile(path)
	if schemaVersionOf(raw) != CurrentSchemaVersion {
		t.Fatalf("upgrade was not written back: %s", raw)
	}
}

func TestMigrateToCurrent_StopsAtAMissingStep(t *testing.T) {
	withMigrations(t, map[int]migrationStep{})

	raw := []byte(`{"discord_app_id": "999"}`)
	got, changed, err := migrateToCurrent(raw)
	if err != nil {
		t.Fatalf("migrateToCurrent: %v", err)
	}
	if changed {
		t.Error("changed = true with no steps registered")
	}
	if string(got) != string(raw) {
		t.Errorf("raw was rewritten: %s", got)
	}
}

func TestLoad_ReportsAFailingStep(t *testing.T) {
	boom := errors.New("boom")
	withMigrations(t, map[int]migrationStep{
		0: func([]byte) ([]byte, error) { return nil, boom },
	})

	writeConfigFile(t, `{"discord_app_id": "999"}`)

	if _, err := Load(); !errors.Is(err, boom) {
		t.Fatalf("Load error = %v, want it to wrap %v", err, boom)
	}
}
