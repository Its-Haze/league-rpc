package livegame

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// fakeDoer routes requests to canned responses by exact URL match.
type fakeDoer struct {
	responses map[string]string
	statusFor map[string]int
	lastURL   string
}

func newFakeDoer() *fakeDoer {
	return &fakeDoer{responses: make(map[string]string), statusFor: make(map[string]int)}
}

func (f *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	f.lastURL = req.URL.String()
	if status, ok := f.statusFor[f.lastURL]; ok {
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	body, ok := f.responses[f.lastURL]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func TestClient_ActivePlayer(t *testing.T) {
	d := newFakeDoer()
	d.responses[baseURL+"/liveclientdata/activeplayer"] = `{"riotId":"Foo#EUW","level":7}`
	c := NewClient(d)

	got, err := c.ActivePlayer(context.Background())
	if err != nil {
		t.Fatalf("ActivePlayer() error = %v", err)
	}
	if got.RiotID != "Foo#EUW" || got.Level != 7 {
		t.Errorf("ActivePlayer() = %+v, want RiotID=Foo#EUW Level=7", got)
	}
}

func TestClient_AllGameData(t *testing.T) {
	d := newFakeDoer()
	d.responses[baseURL+"/liveclientdata/allgamedata"] = `{"allPlayers":[
		{"riotId":"Foo#EUW","rawChampionName":"game_character_displayname_Chogath","rawSkinName":"game_character_skin_displayname_Chogath_1","skinID":1}
	]}`
	c := NewClient(d)

	got, err := c.AllGameData(context.Background())
	if err != nil {
		t.Fatalf("AllGameData() error = %v", err)
	}
	if len(got.AllPlayers) != 1 || got.AllPlayers[0].RiotID != "Foo#EUW" {
		t.Errorf("AllGameData() = %+v", got)
	}
}

func TestClient_GameMode(t *testing.T) {
	d := newFakeDoer()
	d.responses[baseURL+"/liveclientdata/allgamedata"] = `{"gameData":{"gameMode":"PRACTICETOOL"},"allPlayers":[]}`
	c := NewClient(d)

	got, err := c.GameMode(context.Background())
	if err != nil {
		t.Fatalf("GameMode() error = %v", err)
	}
	if got != "PRACTICETOOL" {
		t.Errorf("GameMode() = %q, want PRACTICETOOL", got)
	}
}

func TestClient_PlayerScores(t *testing.T) {
	d := newFakeDoer()
	d.responses[baseURL+"/liveclientdata/playerscores?riotId=Foo%23EUW"] = `{"kills":3,"deaths":2,"assists":5,"creepScore":120}`
	c := NewClient(d)

	got, err := c.PlayerScores(context.Background(), "Foo#EUW")
	if err != nil {
		t.Fatalf("PlayerScores() error = %v", err)
	}
	if got.Kills != 3 || got.Deaths != 2 || got.Assists != 5 || got.CreepScore != 120 {
		t.Errorf("PlayerScores() = %+v", got)
	}
}

func TestClient_GameStats(t *testing.T) {
	d := newFakeDoer()
	d.responses[baseURL+"/liveclientdata/gamestats"] = `{"gameTime":754.321}`
	c := NewClient(d)

	got, err := c.GameStats(context.Background())
	if err != nil {
		t.Fatalf("GameStats() error = %v", err)
	}
	if got.GameTime != 754.321 {
		t.Errorf("GameStats() = %+v, want GameTime=754.321", got)
	}
}

func TestClient_NonOKStatusIsAnError(t *testing.T) {
	d := newFakeDoer()
	d.statusFor[baseURL+"/liveclientdata/activeplayer"] = http.StatusServiceUnavailable
	c := NewClient(d)

	if _, err := c.ActivePlayer(context.Background()); err == nil {
		t.Fatal("ActivePlayer() error = nil, want error on non-200 status")
	}
}

func TestFindPlayer(t *testing.T) {
	players := []Player{{RiotID: "A#1"}, {RiotID: "B#2"}}

	if p, ok := FindPlayer(players, "B#2"); !ok || p.RiotID != "B#2" {
		t.Errorf("FindPlayer(B#2) = %+v, %v", p, ok)
	}
	if _, ok := FindPlayer(players, "C#3"); ok {
		t.Error("FindPlayer(C#3) found a player that isn't in the list")
	}
}
