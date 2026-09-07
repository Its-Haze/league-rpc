package updates

import (
	"os"
	"runtime"
	"strings"
	"testing"

	wupdater "github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

// releaseWorkflow is the pipeline that has to publish ReleaseAsset under that
// exact name for the matcher below to ever see it.
const releaseWorkflow = "../../.github/workflows/release.yml"

func TestReleaseAssetIsSelectedByTheProviderMatcher(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: ChecksumAsset},
		{Name: SignatureAsset},
		{Name: "league-rpc-1.2.3-setup.exe"},
		{Name: ReleaseAsset},
	}
	req := wupdater.CheckRequest{Platform: "windows", Arch: "amd64"}

	got := github.DefaultAssetMatcher(req, assets)
	if got < 0 {
		t.Fatalf("no asset matched; %q must contain the platform and the arch", ReleaseAsset)
	}
	if assets[got].Name != ReleaseAsset {
		t.Fatalf("matched %q, want %q", assets[got].Name, ReleaseAsset)
	}
}

// The updater defaults Platform/Arch to the running build's, so a rename that
// drops "windows" or "amd64" breaks updates on the only platform we ship.
func TestReleaseAssetMatchesTheRunningPlatform(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skipf("release asset targets windows, running on %s", runtime.GOOS)
	}
	req := wupdater.CheckRequest{Platform: runtime.GOOS, Arch: runtime.GOARCH}
	if github.DefaultAssetMatcher(req, []github.ReleaseAsset{{Name: ReleaseAsset}}) < 0 {
		t.Fatalf("%q does not match %s/%s", ReleaseAsset, runtime.GOOS, runtime.GOARCH)
	}
}

func TestReleaseWorkflowPublishesTheAssetName(t *testing.T) {
	body, err := os.ReadFile(releaseWorkflow)
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}
	if !strings.Contains(string(body), ReleaseAsset) {
		t.Fatalf("%s never names %q, so the published asset would not match", releaseWorkflow, ReleaseAsset)
	}
}
