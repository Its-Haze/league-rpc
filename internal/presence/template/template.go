// Package template renders a Discord presence details/state line by plain
// {token} substitution: no logic, no conditionals, deliberately not text/template.
package template

import (
	"maps"
	"regexp"
	"strings"
)

// Context names a presence phase with its own token vocabulary and defaults.
type Context string

const (
	ContextInClient    Context = "in-client"
	ContextLobby       Context = "lobby"
	ContextCustomLobby Context = "custom-lobby"
	ContextQueue       Context = "queue"
	ContextChampSelect Context = "champ-select"
	ContextInGame      Context = "in-game"
	ContextTFTInGame   Context = "tft-in-game"
	ContextSpectating  Context = "spectating"
)

// emptyMark stands in for a known token that resolved to empty, so cleanup can
// tell "author typed nothing here" from "token collapsed to nothing".
const emptyMark = "\x00"

// separators are trimmed when they dangle at an end or pile up after a token
// collapses. U+00B7 is the middot this app's own stat lines use; U+2022 the bullet.
const separators = "-|\u2022\u00b7"

var (
	tokenRe = regexp.MustCompile(`\{([a-zA-Z][a-zA-Z0-9_]*)\}`)
	seamRe  = regexp.MustCompile(`\s*` + emptyMark + `\s*`)
	sepRun  = regexp.MustCompile(`([` + separators + `])(?:\s*[` + separators + `])+`)
	endTrim = regexp.MustCompile(`^[\s` + separators + `]+|[\s` + separators + `]+$`)
)

var knownTokens = map[Context][]string{
	ContextInClient:    {"emoji", "availability"},
	ContextLobby:       {"queue", "players", "max_players"},
	ContextCustomLobby: {"queue", "players", "max_players"},
	ContextQueue:       {"queue"},
	ContextChampSelect: {"queue", "mode"},
	ContextInGame:      {"queue", "mode", "stats", "champion", "skin"},
	ContextTFTInGame:   {"queue", "level"},
	ContextSpectating:  {"mode", "queue"},
}

// defaults are the built-in templates, one [details, state] pair per context.
// Each reproduces the string the app produced before templates were editable.
var defaults = map[Context][2]string{
	ContextInClient:    {"{emoji}  {availability}", "In Client"},
	ContextLobby:       {"{queue}", "In Lobby ({players}/{max_players})"},
	ContextCustomLobby: {"{queue}", "In Lobby"},
	ContextQueue:       {"{queue}", "In Queue"},
	ContextChampSelect: {"{queue}", "In Champ Select"},
	ContextInGame:      {"{queue}", "In Game \u00b7 {stats}"},
	ContextTFTInGame:   {"{queue}", "In Game \u00b7 lvl: {level}"},
	ContextSpectating:  {"{mode}", "Spectating"},
}

var sampleData = map[Context]map[string]string{
	ContextInClient:    {"emoji": "\U0001F7E2", "availability": "Online"},
	ContextLobby:       {"queue": "Ranked Solo/Duo", "players": "3", "max_players": "5"},
	ContextCustomLobby: {"queue": "Practice Tool", "players": "1", "max_players": "1"},
	ContextQueue:       {"queue": "Ranked Solo/Duo"},
	ContextChampSelect: {"queue": "Ranked Solo/Duo", "mode": "Summoner's Rift"},
	ContextInGame: {
		"queue":    "Ranked Solo/Duo",
		"mode":     "Summoner's Rift",
		"stats":    "3/2/5 \u00b7 120cs",
		"champion": "Cho'Gath",
		"skin":     "Battlecast Cho'Gath",
	},
	ContextTFTInGame:  {"queue": "Teamfight Tactics", "level": "7"},
	ContextSpectating: {"mode": "Howling Abyss (ARAM)", "queue": "Howling Abyss (ARAM)"},
}

// Contexts returns every context in a stable, phase-chronological order.
func Contexts() []Context {
	return []Context{
		ContextInClient, ContextLobby, ContextCustomLobby, ContextQueue, ContextChampSelect,
		ContextInGame, ContextTFTInGame, ContextSpectating,
	}
}

// KnownTokens returns the token names valid for ctx, or nil for an unknown ctx.
func KnownTokens(ctx Context) []string {
	src := knownTokens[ctx]
	if src == nil {
		return nil
	}
	out := make([]string, len(src))
	copy(out, src)
	return out
}

// IsContext reports whether ctx is one this package knows.
func IsContext(ctx Context) bool {
	_, ok := knownTokens[ctx]
	return ok
}

// Default returns the built-in details and state templates for ctx.
func Default(ctx Context) (details, state string) {
	d := defaults[ctx]
	return d[0], d[1]
}

// SampleData returns a representative token map for ctx, for previews.
func SampleData(ctx Context) map[string]string {
	return maps.Clone(sampleData[ctx])
}

// RenderPair resolves each template for ctx (the override when non-empty, else
// the default), renders both against data, returns deduped unknown-token names.
func RenderPair(ctx Context, overrideDetails, overrideState string, data map[string]string) (details, state string, warnings []string) {
	detailsTmpl, stateTmpl := Default(ctx)
	if overrideDetails != "" {
		detailsTmpl = overrideDetails
	}
	if overrideState != "" {
		stateTmpl = overrideState
	}

	var dw, sw []string
	details, dw = Render(ctx, detailsTmpl, data)
	state, sw = Render(ctx, stateTmpl, data)

	seen := map[string]bool{}
	for _, name := range append(dw, sw...) {
		if !seen[name] {
			seen[name] = true
			warnings = append(warnings, name)
		}
	}
	return details, state, warnings
}

// Render substitutes each {token} in tmpl from data. Unknown tokens stay literal
// and come back by name; a missing or empty known token collapses with its seam.
func Render(ctx Context, tmpl string, data map[string]string) (string, []string) {
	known := map[string]bool{}
	for _, t := range knownTokens[ctx] {
		known[t] = true
	}

	var unknown []string
	seen := map[string]bool{}

	out := tokenRe.ReplaceAllStringFunc(tmpl, func(m string) string {
		name := m[1 : len(m)-1]
		if !known[name] {
			if !seen[name] {
				seen[name] = true
				unknown = append(unknown, name)
			}
			return m
		}
		if v := data[name]; v != "" {
			return v
		}
		return emptyMark
	})

	return cleanup(out), unknown
}

// cleanup removes empty-token seams, folds runs of separators left behind, and
// trims whitespace or a dangling separator at either end.
func cleanup(s string) string {
	s = collapseSeams(s)
	s = sepRun.ReplaceAllString(s, "$1")
	s = endTrim.ReplaceAllString(s, "")
	return s
}

// collapseSeams replaces each emptyMark and the whitespace around it with a
// single space, or with nothing when the seam sits at the start or end.
func collapseSeams(s string) string {
	locs := seamRe.FindAllStringIndex(s, -1)
	if locs == nil {
		return s
	}
	var b strings.Builder
	prev := 0
	for _, loc := range locs {
		b.WriteString(s[prev:loc[0]])
		if loc[0] != 0 && loc[1] != len(s) {
			b.WriteByte(' ')
		}
		prev = loc[1]
	}
	b.WriteString(s[prev:])
	return b.String()
}
