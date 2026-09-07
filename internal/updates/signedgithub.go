package updates

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	wupdater "github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

// SignatureAsset is the detached ed25519 signature sidecar the release
const SignatureAsset = "SHA256SUMS.sig"

// signedGithubProvider wraps a *github.Provider and fills in the ed25519
type signedGithubProvider struct {
	inner *github.Provider
	repo  string
	doer  HTTPDoer
}

func newSignedGithubProvider(inner *github.Provider, repo string, doer HTTPDoer) *signedGithubProvider {
	return &signedGithubProvider{inner: inner, repo: repo, doer: doer}
}

func (p *signedGithubProvider) Name() string { return p.inner.Name() }

// Check delegates to the wrapped provider, then looks up SignatureAsset on
func (p *signedGithubProvider) Check(ctx context.Context, req wupdater.CheckRequest) (*wupdater.Release, error) {
	rel, err := p.inner.Check(ctx, req)
	if err != nil || rel == nil {
		return rel, err
	}
	sig, err := p.fetchSignature(ctx, rel)
	if err != nil {
		return nil, fmt.Errorf("updates: load release signature: %w", err)
	}
	if sig == nil {
		return rel, nil
	}
	if rel.Verification == nil {
		rel.Verification = &wupdater.Verification{}
	}
	rel.Verification.SignatureAlgo = "ed25519"
	rel.Verification.Signature = sig
	return rel, nil
}

func (p *signedGithubProvider) Download(ctx context.Context, rel *wupdater.Release, dst io.Writer, onProgress func(written, total int64)) error {
	return p.inner.Download(ctx, rel, dst, onProgress)
}

// fetchSignature pulls SignatureAsset from the release tag named in rel's
func (p *signedGithubProvider) fetchSignature(ctx context.Context, rel *wupdater.Release) ([]byte, error) {
	tag, _ := rel.Metadata["github.release.tag"].(string)
	if tag == "" {
		return nil, fmt.Errorf("release metadata missing github.release.tag")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/repos/"+p.repo+"/releases/tags/"+tag, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := p.doer.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("github releases API: HTTP %d: %s", resp.StatusCode, body)
	}

	var payload struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	var sigURL string
	for _, a := range payload.Assets {
		if a.Name == SignatureAsset {
			sigURL = a.BrowserDownloadURL
			break
		}
	}
	if sigURL == "" {
		return nil, nil
	}

	sreq, err := http.NewRequestWithContext(ctx, http.MethodGet, sigURL, nil)
	if err != nil {
		return nil, err
	}
	sreq.Header.Set("Accept", "application/octet-stream")
	sresp, err := p.doer.Do(sreq)
	if err != nil {
		return nil, err
	}
	defer sresp.Body.Close()
	if sresp.StatusCode < 200 || sresp.StatusCode >= 300 {
		return nil, fmt.Errorf("signature sidecar: HTTP %d", sresp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(sresp.Body, 1<<16))
	if err != nil {
		return nil, err
	}
	return parseSignatureLine(string(body), rel.Artifact.Filename)
}

// parseSignatureLine extracts the hex-decoded signature for target from a
// SHA256SUMS.sig-style listing: "<hex-signature>  <filename>" per line.
func parseSignatureLine(body, target string) ([]byte, error) {
	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[len(fields)-1] != target {
			continue
		}
		sig, err := hex.DecodeString(fields[0])
		if err != nil {
			return nil, fmt.Errorf("malformed signature for %s: %w", target, err)
		}
		return sig, nil
	}
	return nil, nil
}
