package updates

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	wupdater "github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

func TestParseSignatureLine(t *testing.T) {
	body := "# comment\ndeadbeef  other.exe\ncafebabe  league-rpc-gui.exe\n"

	got, err := parseSignatureLine(body, "league-rpc-gui.exe")
	if err != nil {
		t.Fatalf("parseSignatureLine: %v", err)
	}
	if hex.EncodeToString(got) != "cafebabe" {
		t.Fatalf("got %x, want cafebabe", got)
	}

	got, err = parseSignatureLine(body, "missing.exe")
	if err != nil || got != nil {
		t.Fatalf("parseSignatureLine(missing) = %x, %v, want nil, nil", got, err)
	}
}

// checksumsFile builds a sha256sum-style listing for one artifact.
func checksumsFile(digest []byte, filename string) string {
	return hex.EncodeToString(digest) + "  " + filename + "\n"
}

func TestSignedProvider_AttachesSignature(t *testing.T) {
	const artifactName = "league-rpc-gui.exe"
	artifactBytes := []byte("pretend-exe-bytes")
	digest := sha256.Sum256(artifactBytes)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	sig := ed25519.Sign(priv, digest[:])

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/repos/owner/repo/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"tag_name": "v1.2.3",
			"assets": [
				{"id": 1, "name": %q, "browser_download_url": %q},
				{"id": 2, "name": "SHA256SUMS", "browser_download_url": %q}
			]
		}`, artifactName, srv.URL+"/dl/"+artifactName, srv.URL+"/dl/SHA256SUMS")
	})
	mux.HandleFunc("/dl/"+artifactName, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(artifactBytes)
	})
	mux.HandleFunc("/dl/SHA256SUMS", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(checksumsFile(digest[:], artifactName)))
	})

	inner, err := github.New(github.Config{
		Repository:    "owner/repo",
		BaseURL:       srv.URL,
		ChecksumAsset: ChecksumAsset,
	})
	if err != nil {
		t.Fatalf("github.New: %v", err)
	}

	doer := doerFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.String() {
		case "https://api.github.com/repos/owner/repo/releases/tags/v1.2.3":
			return jsonResponse(200, fmt.Sprintf(`{"assets":[{"name":%q,"browser_download_url":"https://sig.test/SHA256SUMS.sig"}]}`, SignatureAsset)), nil
		case "https://sig.test/SHA256SUMS.sig":
			return jsonResponse(200, checksumsFile(sig, artifactName)), nil
		}
		return nil, fmt.Errorf("unexpected request to %s", r.URL)
	})

	p := newSignedGithubProvider(inner, "owner/repo", doer)

	rel, err := p.Check(context.Background(), updaterCheckRequest("0.0.0"))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if rel == nil {
		t.Fatal("Check returned nil release, want an update")
	}
	if rel.Verification == nil {
		t.Fatal("Verification is nil, want digest+signature")
	}
	if rel.Verification.SignatureAlgo != "ed25519" {
		t.Fatalf("SignatureAlgo = %q, want ed25519", rel.Verification.SignatureAlgo)
	}
	if !ed25519.Verify(pub, digest[:], rel.Verification.Signature) {
		t.Fatal("attached signature does not verify against the digest")
	}
}

func TestSignedProvider_NoSidecarLeavesDigestOnly(t *testing.T) {
	const artifactName = "league-rpc-gui.exe"
	artifactBytes := []byte("pretend-exe-bytes")
	digest := sha256.Sum256(artifactBytes)

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/repos/owner/repo/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"tag_name": "v1.2.3",
			"assets": [
				{"id": 1, "name": %q, "browser_download_url": %q},
				{"id": 2, "name": "SHA256SUMS", "browser_download_url": %q}
			]
		}`, artifactName, srv.URL+"/dl/"+artifactName, srv.URL+"/dl/SHA256SUMS")
	})
	mux.HandleFunc("/dl/"+artifactName, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(artifactBytes)
	})
	mux.HandleFunc("/dl/SHA256SUMS", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(checksumsFile(digest[:], artifactName)))
	})

	inner, err := github.New(github.Config{
		Repository:    "owner/repo",
		BaseURL:       srv.URL,
		ChecksumAsset: ChecksumAsset,
	})
	if err != nil {
		t.Fatalf("github.New: %v", err)
	}

	// A release with no SHA256SUMS.sig asset at all: an older release published
	// before this pipeline existed.
	doer := doerFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"assets":[]}`), nil
	})

	p := newSignedGithubProvider(inner, "owner/repo", doer)
	rel, err := p.Check(context.Background(), updaterCheckRequest("0.0.0"))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if rel == nil || rel.Verification == nil {
		t.Fatal("expected a digest-only Verification")
	}
	if len(rel.Verification.Signature) != 0 {
		t.Fatal("expected no signature when no sidecar is published")
	}
}

// updaterCheckRequest builds a CheckRequest with a current version old enough
// that the fixed "v1.2.3" fixture release always looks newer.
func updaterCheckRequest(currentVersion string) wupdater.CheckRequest {
	return wupdater.CheckRequest{CurrentVersion: currentVersion}
}
