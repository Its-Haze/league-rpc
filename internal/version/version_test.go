package version

import "testing"

func TestResolve(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty is dev placeholder", "", DevPlaceholder},
		{"whitespace is dev placeholder", "   ", DevPlaceholder},
		{"plain semver passes through", "1.2.3", "1.2.3"},
		{"leading v is trimmed", "v1.2.3", "1.2.3"},
		{"prerelease tag kept", "v2.0.0-rc.1", "2.0.0-rc.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolve(tt.raw); got != tt.want {
				t.Fatalf("resolve(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestVersionAndIsDev_DefaultBuild(t *testing.T) {
	// The test binary carries no -ldflags injection.
	if !IsDev() {
		t.Fatal("IsDev() = false for an un-injected build, want true")
	}
	if got := Version(); got != DevPlaceholder {
		t.Fatalf("Version() = %q, want %q", got, DevPlaceholder)
	}
}
