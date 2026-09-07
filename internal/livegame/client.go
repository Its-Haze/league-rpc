// Package livegame talks to the unauthenticated Live Client Data API at
package livegame

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	baseURL        = "https://127.0.0.1:2999"
	requestTimeout = 3 * time.Second
)

// HTTPDoer is anything that can execute an *http.Request, satisfied by
// *http.Client. Injected so tests never hit a real League client.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// ActivePlayer is the response shape of /liveclientdata/activeplayer.
type ActivePlayer struct {
	RiotID      string  `json:"riotId"`
	Level       int     `json:"level"`
	CurrentGold float64 `json:"currentGold"`
}

// Player is one entry in AllGameData's allPlayers list.
type Player struct {
	RiotID          string `json:"riotId"`
	RawChampionName string `json:"rawChampionName"`
	RawSkinName     string `json:"rawSkinName"`
	SkinID          int    `json:"skinID"`
}

// GameData is the "gameData" object in AllGameData's response.
type GameData struct {
	GameMode  string  `json:"gameMode"`
	GameTime  float64 `json:"gameTime"`
	MapNumber int     `json:"mapNumber"`
}

// AllGameData is the response shape of /liveclientdata/allgamedata.
type AllGameData struct {
	GameData   GameData `json:"gameData"`
	AllPlayers []Player `json:"allPlayers"`
}

// PlayerScores is the response shape of /liveclientdata/playerscores.
type PlayerScores struct {
	Kills      int `json:"kills"`
	Deaths     int `json:"deaths"`
	Assists    int `json:"assists"`
	CreepScore int `json:"creepScore"`
}

// GameStats is the response shape of /liveclientdata/gamestats.
type GameStats struct {
	GameTime float64 `json:"gameTime"`
}

// Client issues Live Client Data requests through an injected HTTPDoer.
type Client struct {
	doer HTTPDoer
}

// NewClient builds a Client that issues requests through doer.
func NewClient(doer HTTPDoer) *Client {
	return &Client{doer: doer}
}

// NewProductionHTTPDoer returns an *http.Client configured for Live Client
// Data's self-signed certificate, matching the LCU client's own TLS setup.
func NewProductionHTTPDoer() HTTPDoer {
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec // self-signed local cert, same as lcu.Client
		Timeout:   requestTimeout,
	}
}

// ActivePlayer polls /liveclientdata/activeplayer.
func (c *Client) ActivePlayer(ctx context.Context) (ActivePlayer, error) {
	var ap ActivePlayer
	err := c.getJSON(ctx, baseURL+"/liveclientdata/activeplayer", &ap)
	return ap, err
}

// AllGameData polls /liveclientdata/allgamedata.
func (c *Client) AllGameData(ctx context.Context) (AllGameData, error) {
	var d AllGameData
	err := c.getJSON(ctx, baseURL+"/liveclientdata/allgamedata", &d)
	return d, err
}

// PlayerScores polls /liveclientdata/playerscores for riotID.
func (c *Client) PlayerScores(ctx context.Context, riotID string) (PlayerScores, error) {
	var s PlayerScores
	u := baseURL + "/liveclientdata/playerscores?riotId=" + url.QueryEscape(riotID)
	err := c.getJSON(ctx, u, &s)
	return s, err
}

// GameStats polls /liveclientdata/gamestats.
func (c *Client) GameStats(ctx context.Context) (GameStats, error) {
	var g GameStats
	err := c.getJSON(ctx, baseURL+"/liveclientdata/gamestats", &g)
	return g, err
}

// GameMode fetches allgamedata and returns just the raw LCU-style gameMode
func (c *Client) GameMode(ctx context.Context) (string, error) {
	data, err := c.AllGameData(ctx)
	if err != nil {
		return "", err
	}
	return data.GameData.GameMode, nil
}

// FindPlayer returns the entry in players whose riotId matches riotID.
func FindPlayer(players []Player, riotID string) (Player, bool) {
	for _, p := range players {
		if p.RiotID == riotID {
			return p, true
		}
	}
	return Player{}, false
}

func (c *Client) getJSON(ctx context.Context, u string, out any) error {
	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("livegame: unexpected status %d from %s", resp.StatusCode, u)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
