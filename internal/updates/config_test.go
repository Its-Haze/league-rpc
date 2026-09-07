package updates

import "testing"

// This guards against a real regression: publicKeyPEM is populated by a
func TestBuildConfig_EmbedsNonEmptyPublicKey(t *testing.T) {
	cfg, err := BuildConfig("1.0.0")
	if err != nil {
		t.Fatalf("BuildConfig: %v", err)
	}
	if len(cfg.PublicKey) == 0 {
		t.Fatal("Config.PublicKey is empty; //go:embed keys/update-public.pem did not populate publicKeyPEM")
	}
}
