package template

import (
	"reflect"
	"testing"
)

func TestRender_Substitution(t *testing.T) {
	tests := []struct {
		name string
		ctx  Context
		tmpl string
		data map[string]string
		want string
	}{
		{
			name: "single token",
			ctx:  ContextChampSelect,
			tmpl: "{queue}",
			data: map[string]string{"queue": "Ranked Solo/Duo"},
			want: "Ranked Solo/Duo",
		},
		{
			name: "token among literal text",
			ctx:  ContextInGame,
			tmpl: "In Game · {stats}",
			data: map[string]string{"stats": "3/2/5 · 120cs"},
			want: "In Game · 3/2/5 · 120cs",
		},
		{
			name: "repeated token substituted every time",
			ctx:  ContextSpectating,
			tmpl: "{mode} — watching {mode}",
			data: map[string]string{"mode": "ARAM"},
			want: "ARAM — watching ARAM",
		},
		{
			name: "literal two spaces between tokens are preserved",
			ctx:  ContextInClient,
			tmpl: "{emoji}  {availability}",
			data: map[string]string{"emoji": "🟢", "availability": "Online"},
			want: "🟢  Online",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warnings := Render(tt.ctx, tt.tmpl, tt.data)
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
			if len(warnings) != 0 {
				t.Errorf("warnings = %v, want none", warnings)
			}
		})
	}
}

func TestRender_UnknownTokenLeftLiteralAndReported(t *testing.T) {
	got, warnings := Render(ContextInClient, "hi {availability} {foo} {bar} {foo}", map[string]string{
		"availability": "Online",
	})
	if want := "hi Online {foo} {bar} {foo}"; got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
	if want := []string{"foo", "bar"}; !reflect.DeepEqual(warnings, want) {
		t.Errorf("warnings = %v, want %v (first appearance, deduped)", warnings, want)
	}
}

func TestRender_EmptyTokenCollapse(t *testing.T) {
	tests := []struct {
		name string
		ctx  Context
		tmpl string
		data map[string]string
		want string
	}{
		{
			name: "leading empty token and its whitespace vanish",
			ctx:  ContextInClient,
			tmpl: "{emoji}  {availability}",
			data: map[string]string{"availability": "Online"},
			want: "Online",
		},
		{
			name: "trailing empty token leaves no space",
			ctx:  ContextInGame,
			tmpl: "In Game {stats}",
			data: map[string]string{},
			want: "In Game",
		},
		{
			name: "trailing empty token with dangling middot",
			ctx:  ContextInGame,
			tmpl: "In Game · {stats}",
			data: map[string]string{},
			want: "In Game",
		},
		{
			name: "middle empty token collapses to one space",
			ctx:  ContextInGame,
			tmpl: "{champion} {stats} {skin}",
			data: map[string]string{"champion": "Ahri", "skin": "Spirit Blossom Ahri"},
			want: "Ahri Spirit Blossom Ahri",
		},
		{
			name: "middle empty token between separators folds the run",
			ctx:  ContextInGame,
			tmpl: "{champion} • {stats} • {skin}",
			data: map[string]string{"champion": "Ahri", "skin": "Spirit Blossom Ahri"},
			want: "Ahri • Spirit Blossom Ahri",
		},
		{
			name: "everything empty renders empty",
			ctx:  ContextSpectating,
			tmpl: "{mode}",
			data: map[string]string{},
			want: "",
		},
		{
			name: "non-empty middot separators inside a value are untouched",
			ctx:  ContextInGame,
			tmpl: "In Game · {stats}",
			data: map[string]string{"stats": "4/1/7 · lvl: 14 · gold: 8200"},
			want: "In Game · 4/1/7 · lvl: 14 · gold: 8200",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := Render(tt.ctx, tt.tmpl, tt.data)
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Every default template must reproduce the exact string the app rendered
// before templates existed.
func TestDefault_ReproducesLegacyStrings(t *testing.T) {
	tests := []struct {
		name                   string
		ctx                    Context
		data                   map[string]string
		wantDetails, wantState string
	}{
		{
			name:        "in-client, emoji on, online",
			ctx:         ContextInClient,
			data:        map[string]string{"emoji": "🟢", "availability": "Online"},
			wantDetails: "🟢  Online",
			wantState:   "In Client",
		},
		{
			name:        "in-client, emoji on, away",
			ctx:         ContextInClient,
			data:        map[string]string{"emoji": "🔴", "availability": "Away"},
			wantDetails: "🔴  Away",
			wantState:   "In Client",
		},
		{
			name:        "in-client, emoji off",
			ctx:         ContextInClient,
			data:        map[string]string{"availability": "Online"},
			wantDetails: "Online",
			wantState:   "In Client",
		},
		{
			name:        "lobby, full party",
			ctx:         ContextLobby,
			data:        map[string]string{"queue": "Ranked Solo/Duo", "players": "3", "max_players": "5"},
			wantDetails: "Ranked Solo/Duo",
			wantState:   "In Lobby (3/5)",
		},
		{
			name:        "custom-lobby",
			ctx:         ContextCustomLobby,
			data:        map[string]string{"queue": "Practice Tool", "players": "1", "max_players": "1"},
			wantDetails: "Practice Tool",
			wantState:   "In Lobby",
		},
		{
			name:        "queue",
			ctx:         ContextQueue,
			data:        map[string]string{"queue": "Ranked Solo/Duo"},
			wantDetails: "Ranked Solo/Duo",
			wantState:   "In Queue",
		},
		{
			name:        "champ-select",
			ctx:         ContextChampSelect,
			data:        map[string]string{"queue": "Ranked Solo/Duo"},
			wantDetails: "Ranked Solo/Duo",
			wantState:   "In Champ Select",
		},
		{
			name:        "in-game, stats shown",
			ctx:         ContextInGame,
			data:        map[string]string{"queue": "Ranked Solo/Duo", "stats": "3/2/5 · 120cs"},
			wantDetails: "Ranked Solo/Duo",
			wantState:   "In Game · 3/2/5 · 120cs",
		},
		{
			name:        "in-game, stats hidden",
			ctx:         ContextInGame,
			data:        map[string]string{"queue": "Ranked Solo/Duo"},
			wantDetails: "Ranked Solo/Duo",
			wantState:   "In Game",
		},
		{
			name:        "in-game, arena stat line",
			ctx:         ContextInGame,
			data:        map[string]string{"queue": "Arena", "stats": "4/1/7 · lvl: 14 · gold: 8200"},
			wantDetails: "Arena",
			wantState:   "In Game · 4/1/7 · lvl: 14 · gold: 8200",
		},
		{
			name:        "tft-in-game",
			ctx:         ContextTFTInGame,
			data:        map[string]string{"queue": "Teamfight Tactics", "level": "7"},
			wantDetails: "Teamfight Tactics",
			wantState:   "In Game · lvl: 7",
		},
		{
			name:        "spectating, mode resolved",
			ctx:         ContextSpectating,
			data:        map[string]string{"mode": "Howling Abyss (ARAM)"},
			wantDetails: "Howling Abyss (ARAM)",
			wantState:   "Spectating",
		},
		{
			name:        "spectating, mode unresolved renders empty details",
			ctx:         ContextSpectating,
			data:        map[string]string{},
			wantDetails: "",
			wantState:   "Spectating",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, st := Default(tt.ctx)
			gotDetails, _ := Render(tt.ctx, dt, tt.data)
			gotState, _ := Render(tt.ctx, st, tt.data)
			if gotDetails != tt.wantDetails {
				t.Errorf("details = %q, want %q", gotDetails, tt.wantDetails)
			}
			if gotState != tt.wantState {
				t.Errorf("state = %q, want %q", gotState, tt.wantState)
			}
		})
	}
}

func TestRenderPair_BlankOverrideFallsBackToDefault(t *testing.T) {
	data := map[string]string{"queue": "Ranked Solo/Duo", "stats": "3/2/5 · 120cs"}

	// Both lines blank: identical to rendering the two defaults.
	d, s, warn := RenderPair(ContextInGame, "", "", data)
	if d != "Ranked Solo/Duo" || s != "In Game · 3/2/5 · 120cs" {
		t.Fatalf("blank override = %q / %q, want the defaults", d, s)
	}
	if warn != nil {
		t.Fatalf("warnings = %v, want none", warn)
	}

	// One line overridden, the other blank and still defaulted.
	d, s, _ = RenderPair(ContextInGame, "{queue} ranked", "", data)
	if d != "Ranked Solo/Duo ranked" || s != "In Game · 3/2/5 · 120cs" {
		t.Fatalf("mixed override = %q / %q", d, s)
	}
}

func TestRenderPair_DedupesUnknownAcrossLines(t *testing.T) {
	_, _, warn := RenderPair(ContextInClient, "{foo} {availability}", "{foo} {bar}", map[string]string{"availability": "Online"})
	if want := []string{"foo", "bar"}; !reflect.DeepEqual(warn, want) {
		t.Fatalf("warnings = %v, want %v", warn, want)
	}
}

func TestContextsAndKnownTokens(t *testing.T) {
	if len(Contexts()) != 8 {
		t.Fatalf("Contexts() = %v, want 8", Contexts())
	}
	for _, ctx := range Contexts() {
		if !IsContext(ctx) {
			t.Errorf("IsContext(%q) = false", ctx)
		}
		if len(KnownTokens(ctx)) == 0 {
			t.Errorf("KnownTokens(%q) is empty", ctx)
		}
	}
	if IsContext("nonsense") {
		t.Error("IsContext(nonsense) = true")
	}
	if KnownTokens("nonsense") != nil {
		t.Error("KnownTokens(nonsense) should be nil")
	}
}
