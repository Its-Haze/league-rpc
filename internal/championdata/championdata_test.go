package championdata

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

// fakeDoer routes requests to canned responses by exact URL match, and
// counts how many times each URL was requested (to assert caching).
type fakeDoer struct {
	responses map[string]string // url -> JSON body
	notFound  map[string]bool   // url -> should 404
	calls     map[string]*atomic.Int32
}

func newFakeDoer() *fakeDoer {
	return &fakeDoer{
		responses: make(map[string]string),
		notFound:  make(map[string]bool),
		calls:     make(map[string]*atomic.Int32),
	}
}

func (f *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	if f.calls[url] == nil {
		f.calls[url] = &atomic.Int32{}
	}
	f.calls[url].Add(1)

	if f.notFound[url] {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil
	}

	body, ok := f.responses[url]
	if !ok {
		return nil, errors.New("fakeDoer: no canned response for " + url)
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func (f *fakeDoer) callCount(url string) int32 {
	if c := f.calls[url]; c != nil {
		return c.Load()
	}
	return 0
}

const versionsBody = `["14.1.1"]`

func chogathChampionURL() string {
	return "https://ddragon.leagueoflegends.com/cdn/14.1.1/data/en_US/champion/Chogath.json"
}

func chogathChampionBody() string {
	payload := championPayload{
		Data: map[string]championEntry{
			"Chogath": {
				ID:   "Chogath",
				Name: "Cho'Gath",
				Skins: []skinEntry{
					{Num: 0, Name: "default"},
					{Num: 1, Name: "Battlecast Cho'Gath"},
					{Num: 17, Name: "Corporate Mundo (Rose Quartz)"}, // chroma-shaped entry, should never be picked as base
				},
			},
		},
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

func newDoerWithChogath() *fakeDoer {
	d := newFakeDoer()
	d.responses["https://ddragon.leagueoflegends.com/api/versions.json"] = versionsBody
	d.responses[chogathChampionURL()] = chogathChampionBody()
	d.responses["https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/champions.json"] = `{}`
	return d
}

func TestResolve_PrimaryRawChampionName(t *testing.T) {
	d := newDoerWithChogath()
	r := NewResolver(d)

	got, err := r.Resolve(context.Background(), "game_character_displayname_Chogath", "game_character_skin_displayname_Chogath_1", 1)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.ID != "Chogath" || got.Name != "Cho'Gath" {
		t.Errorf("Resolve() ID/Name = %q/%q, want Chogath/Cho'Gath", got.ID, got.Name)
	}
	if got.SkinName != "Battlecast Cho'Gath" {
		t.Errorf("SkinName = %q, want Battlecast Cho'Gath", got.SkinName)
	}
	if got.ChromaName != "" {
		t.Errorf("ChromaName = %q, want empty", got.ChromaName)
	}
}

func TestResolve_DefaultSkinFallback(t *testing.T) {
	// rawChampionName is the "Character_Chogath_Name" placeholder format
	d := newDoerWithChogath()
	r := NewResolver(d)

	got, err := r.Resolve(context.Background(), "Character_Chogath_Name", "game_character_displayname_Chogath", 0)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.ID != "Chogath" {
		t.Errorf("Resolve() ID = %q, want Chogath", got.ID)
	}
	if got.SkinName != "" {
		t.Errorf("SkinName = %q, want empty for default skin", got.SkinName)
	}
}

func TestResolve_KDAVariantFallback(t *testing.T) {
	d := newFakeDoer()
	d.responses["https://ddragon.leagueoflegends.com/api/versions.json"] = versionsBody
	d.responses["https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/champions.json"] = `{}`

	seraphinePayload := championPayload{
		Data: map[string]championEntry{
			"Seraphine": {
				ID:   "Seraphine",
				Name: "Seraphine",
				Skins: []skinEntry{
					{Num: 0, Name: "default"},
					{Num: 1, Name: "K/DA ALL OUT Seraphine"},
				},
			},
		},
	}
	b, _ := json.Marshal(seraphinePayload)
	d.responses["https://ddragon.leagueoflegends.com/cdn/14.1.1/data/en_US/champion/Seraphine.json"] = string(b)
	// Primary and default-skin candidates 404, forcing the K/DA-variant fallback.
	d.notFound["https://ddragon.leagueoflegends.com/cdn/14.1.1/data/en_US/champion/KDA.json"] = true

	r := NewResolver(d)
	got, err := r.Resolve(context.Background(),
		"game_character_displayname_Seraphine_KDA",
		"game_character_skin_displayname_Seraphine_1",
		1,
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.ID != "Seraphine" {
		t.Errorf("Resolve() ID = %q, want Seraphine", got.ID)
	}
	if got.SkinName != "K/DA ALL OUT Seraphine" {
		t.Errorf("SkinName = %q, want K/DA ALL OUT Seraphine", got.SkinName)
	}
}

func TestResolve_LoadingScreenPlaceholder_ReturnsErrUnresolved(t *testing.T) {
	d := newFakeDoer()
	d.responses["https://ddragon.leagueoflegends.com/api/versions.json"] = versionsBody

	r := NewResolver(d)
	_, err := r.Resolve(context.Background(), "Character_Unknown_Name", "", 0)
	if !errors.Is(err, ErrUnresolved) {
		t.Errorf("Resolve() error = %v, want ErrUnresolved", err)
	}
}

func TestResolve_ChromaResolvedViaMeraki(t *testing.T) {
	d := newDoerWithChogath()

	meraki := map[string]merakiChampion{
		"Chogath": {
			Skins: []merakiSkin{
				{
					ID: 36001, // base skin num 1 (Battlecast)
					Chromas: []merakiChroma{
						{ID: 36004, Name: "Emerald"}, // 36004 % 1000 == 4
					},
				},
			},
		},
	}
	b, _ := json.Marshal(meraki)
	d.responses["https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/champions.json"] = string(b)

	r := NewResolver(d)
	got, err := r.Resolve(context.Background(), "game_character_displayname_Chogath", "game_character_skin_displayname_Chogath_4", 4)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.BaseSkinID != 1 {
		t.Errorf("BaseSkinID = %d, want 1 (parent of the chroma)", got.BaseSkinID)
	}
	if got.ChromaName != "Emerald" {
		t.Errorf("ChromaName = %q, want Emerald", got.ChromaName)
	}
	if got.SkinName != "Battlecast Cho'Gath" {
		t.Errorf("SkinName = %q, want Battlecast Cho'Gath (the chroma's base skin)", got.SkinName)
	}
}

func TestResolve_DDragonHeuristicFallbackWhenMerakiHasNoMatch(t *testing.T) {
	// skinID 18 isn't a real DDragon num for this champion (only 0, 1, 17
	d := newDoerWithChogath()
	r := NewResolver(d)

	got, err := r.Resolve(context.Background(), "game_character_displayname_Chogath", "game_character_skin_displayname_Chogath_18", 18)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.BaseSkinID != 1 {
		t.Errorf("BaseSkinID = %d, want 1 (highest real base skin <= 18)", got.BaseSkinID)
	}
}

func TestResolve_CachesChampionJSONAndVersion(t *testing.T) {
	d := newDoerWithChogath()
	r := NewResolver(d)

	for range 3 {
		if _, err := r.Resolve(context.Background(), "game_character_displayname_Chogath", "game_character_skin_displayname_Chogath_1", 1); err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	}

	if got := d.callCount("https://ddragon.leagueoflegends.com/api/versions.json"); got != 1 {
		t.Errorf("versions.json fetched %d times, want 1 (cached)", got)
	}
	if got := d.callCount(chogathChampionURL()); got != 1 {
		t.Errorf("champion JSON fetched %d times, want 1 (cached)", got)
	}
}
