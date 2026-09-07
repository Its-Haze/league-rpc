package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_RoundTripsASavedConfig(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())

	orig := DefaultConfig()
	orig.Theme = ThemeDark
	orig.Display.Default.ShowRank = false
	if err := Save(orig); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", got.SchemaVersion, CurrentSchemaVersion)
	}

	a, _ := json.Marshal(orig)
	b, _ := json.Marshal(got)
	if string(a) != string(b) {
		t.Fatalf("config did not round-trip:\n%s\n%s", a, b)
	}
}

func TestLoad_BrokenFileClampsToValid(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)

	path, _ := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	broken := `{"schema_version": 1, "discord_app_id": "", "theme": "chartreuse",
	  "advanced": {"update_interval": 1, "stats_polling_interval": 9999999}}`
	if err := os.WriteFile(path, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("Load returned an invalid config: %v", err)
	}
}
