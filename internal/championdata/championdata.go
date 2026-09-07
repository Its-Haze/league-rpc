// Package championdata resolves a raw champion/skin identifier from Live
package championdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	ddragonVersionsURL    = "https://ddragon.leagueoflegends.com/api/versions.json"
	ddragonChampionURLFmt = "https://ddragon.leagueoflegends.com/cdn/%s/data/%s/champion/%s.json"
	merakiChampionsURL    = "https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/champions.json"

	ddragonLocale = "en_US"

	requestTimeout = 5 * time.Second
)

// ErrUnresolved means none of the raw-name candidates resolved to a known
var ErrUnresolved = errors.New("championdata: could not resolve champion from raw names")

// HTTPDoer is anything that can execute an *http.Request, satisfied by
// *http.Client. Injected so tests never hit real Data Dragon/Meraki servers.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Resolution is a fully resolved champion/skin/chroma identity.
type Resolution struct {
	ID         string // Data Dragon's URL-safe champion id, e.g. "Chogath"
	Name       string // Data Dragon's display name, e.g. "Cho'Gath"
	BaseSkinID int    // The resolved base skin's DDragon num (feeds the tile URL)
	SkinName   string // Empty for the default skin
	ChromaName string // Empty when skinID isn't a chroma
}

type championPayload struct {
	Data map[string]championEntry `json:"data"`
}

type championEntry struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Skins []skinEntry `json:"skins"`
}

type skinEntry struct {
	Num  int    `json:"num"`
	Name string `json:"name"`
}

type merakiChampion struct {
	Skins []merakiSkin `json:"skins"`
}

type merakiSkin struct {
	ID      int            `json:"id"`
	Chromas []merakiChroma `json:"chromas"`
}

type merakiChroma struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Resolver resolves raw champion identifiers via Data Dragon and Meraki,
type Resolver struct {
	doer HTTPDoer

	mu      sync.Mutex
	version string
	champs  map[string]championPayload // keyed by "<name>|<version>"
	meraki  map[string]merakiChampion
}

// NewResolver builds a Resolver that issues requests through doer.
func NewResolver(doer HTTPDoer) *Resolver {
	return &Resolver{
		doer:   doer,
		champs: make(map[string]championPayload),
	}
}

// NewProductionHTTPDoer returns an *http.Client suitable for real Data
func NewProductionHTTPDoer() HTTPDoer {
	return &http.Client{Timeout: requestTimeout}
}

// Resolve turns a raw champion/skin/chroma triple from Live Client Data into
func (r *Resolver) Resolve(ctx context.Context, rawChampionName, rawSkinName string, skinID int) (Resolution, error) {
	var (
		champ   championEntry
		found   bool
		lastErr error
	)

	for _, candidate := range candidateNames(rawChampionName, rawSkinName) {
		if !validCandidate(candidate) {
			continue
		}
		payload, err := r.getChampionPayload(ctx, candidate)
		if err != nil {
			lastErr = err
			continue
		}
		entry, ok := payload.Data[candidate]
		if !ok {
			continue
		}
		champ = entry
		found = true
		break
	}

	if !found {
		if lastErr != nil {
			return Resolution{}, lastErr
		}
		return Resolution{}, ErrUnresolved
	}

	baseSkinID, chromaName := r.resolveSkin(ctx, champ.ID, skinID, champ.Skins)

	return Resolution{
		ID:         champ.ID,
		Name:       champ.Name,
		BaseSkinID: baseSkinID,
		SkinName:   skinNameForNum(champ.Skins, baseSkinID),
		ChromaName: chromaName,
	}, nil
}

// candidateNames returns the three-tier raw-name fallback chain: the
func candidateNames(rawChampionName, rawSkinName string) []string {
	var out []string
	if v := lastSegment(rawChampionName, 1); v != "" {
		out = append(out, v)
	}
	if v := lastSegment(rawSkinName, 1); v != "" {
		out = append(out, v)
	}
	if v := lastSegment(rawSkinName, 2); v != "" {
		out = append(out, v)
	}
	return out
}

// lastSegment splits s on "_" and returns the nth element from the end
// (n=1 is the last element), or "" if s has fewer than n segments.
func lastSegment(s string, n int) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	if len(parts) < n {
		return ""
	}
	return parts[len(parts)-n]
}

// validCandidate rejects the loading-screen placeholder values the live
// game API returns before champion data is actually available.
func validCandidate(name string) bool {
	switch name {
	case "", "Name", "Unknown":
		return false
	default:
		return true
	}
}

// resolveSkin finds skinID's base skin num and, if skinID is itself a
func (r *Resolver) resolveSkin(ctx context.Context, champKey string, skinID int, ddragonSkins []skinEntry) (int, string) {
	if meraki, err := r.getMerakiChampion(ctx, champKey); err == nil {
		for _, mskin := range meraki.Skins {
			for _, chroma := range mskin.Chromas {
				if chroma.ID%1000 == skinID {
					return mskin.ID % 1000, chroma.Name
				}
			}
		}
	}

	for _, skin := range ddragonSkins {
		if skin.Num == skinID && isBaseSkinName(skin.Name) {
			return skinID, ""
		}
	}

	best := -1
	for _, skin := range ddragonSkins {
		if skin.Num <= skinID && skin.Num > best && isBaseSkinName(skin.Name) {
			best = skin.Num
		}
	}
	if best >= 0 {
		return best, ""
	}

	return 0, ""
}

// isBaseSkinName reports whether a DDragon skin entry's name looks like a
// base skin rather than a chroma (chromas are named "Skin Name (Color)").
func isBaseSkinName(name string) bool {
	return name == "default" || !strings.Contains(name, "(")
}

func skinNameForNum(skins []skinEntry, num int) string {
	for _, s := range skins {
		if s.Num == num {
			if s.Name == "default" {
				return ""
			}
			return s.Name
		}
	}
	return ""
}

func (r *Resolver) getVersion(ctx context.Context) (string, error) {
	r.mu.Lock()
	if r.version != "" {
		v := r.version
		r.mu.Unlock()
		return v, nil
	}
	r.mu.Unlock()

	var versions []string
	if err := r.getJSON(ctx, ddragonVersionsURL, &versions); err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", errors.New("championdata: versions.json returned no versions")
	}

	r.mu.Lock()
	r.version = versions[0]
	v := r.version
	r.mu.Unlock()
	return v, nil
}

func (r *Resolver) getChampionPayload(ctx context.Context, name string) (championPayload, error) {
	version, err := r.getVersion(ctx)
	if err != nil {
		return championPayload{}, err
	}

	key := name + "|" + version
	r.mu.Lock()
	if payload, ok := r.champs[key]; ok {
		r.mu.Unlock()
		return payload, nil
	}
	r.mu.Unlock()

	url := fmt.Sprintf(ddragonChampionURLFmt, version, ddragonLocale, name)
	var payload championPayload
	if err := r.getJSON(ctx, url, &payload); err != nil {
		return championPayload{}, err
	}

	r.mu.Lock()
	r.champs[key] = payload
	r.mu.Unlock()
	return payload, nil
}

func (r *Resolver) getMerakiChampion(ctx context.Context, champKey string) (merakiChampion, error) {
	r.mu.Lock()
	meraki := r.meraki
	r.mu.Unlock()

	if meraki == nil {
		var fetched map[string]merakiChampion
		if err := r.getJSON(ctx, merakiChampionsURL, &fetched); err != nil {
			return merakiChampion{}, err
		}
		r.mu.Lock()
		r.meraki = fetched
		meraki = r.meraki
		r.mu.Unlock()
	}

	champ, ok := meraki[champKey]
	if !ok {
		return merakiChampion{}, fmt.Errorf("championdata: %q not found in meraki data", champKey)
	}
	return champ, nil
}

func (r *Resolver) getJSON(ctx context.Context, url string, out any) error {
	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := r.doer.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("championdata: unexpected status %d from %s", resp.StatusCode, url)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
