package discord

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Presence images are hotlinked out of this repository's own assets/ folder,
// so deleting or renaming one silently breaks every user's Discord presence.
func TestLogoURLsPointAtCommittedAssets(t *testing.T) {
	const prefix = "https://github.com/Its-Haze/league-rpc/blob/master/"

	for name, url := range map[string]string{
		"leagueLogoURL":      leagueLogoURL,
		"leagueLogoLargeURL": leagueLogoLargeURL,
	} {
		rel, ok := strings.CutPrefix(url, prefix)
		if !ok {
			t.Errorf("%s = %q, want it to start with %q; update this test if the assets moved host", name, url, prefix)
			continue
		}
		rel = strings.TrimSuffix(rel, "?raw=true")
		path := filepath.Join("..", "..", filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s points at %q, which is not committed: %v", name, rel, err)
		}
	}
}
