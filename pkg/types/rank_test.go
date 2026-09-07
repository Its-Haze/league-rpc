package types

import "testing"

func TestRankedStats_String(t *testing.T) {
	tests := []struct {
		name string
		rank RankedStats
		want string
	}{
		{"unranked", RankedStats{}, "Unranked"},
		{"division", RankedStats{Tier: TierGold, Division: DivisionIV, LeaguePoints: 57}, "Gold IV: 57 LP"},
		{"master has no division", RankedStats{Tier: TierMaster, Division: DivisionNA, LeaguePoints: 812}, "Master: 812 LP"},
		{"grandmaster", RankedStats{Tier: TierGrandmaster, LeaguePoints: 300}, "Grandmaster: 300 LP"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rank.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTier_DisplayName(t *testing.T) {
	tests := []struct {
		tier Tier
		want string
	}{
		{TierGold, "Gold"},
		{TierGrandmaster, "Grandmaster"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := tt.tier.DisplayName(); got != tt.want {
			t.Errorf("DisplayName(%q) = %q, want %q", tt.tier, got, tt.want)
		}
	}
}

func TestMapRatedTierToTier(t *testing.T) {
	tests := []struct {
		rated ArenaRatedTier
		want  ArenaTier
	}{
		{ArenaRatedNone, ""},
		{ArenaRatedGray, ArenaTierWood},
		{ArenaRatedGreen, ArenaTierBronze},
		{ArenaRatedBlue, ArenaTierSilver},
		{ArenaRatedPurple, ArenaTierGold},
		{ArenaRatedOrange, ArenaTierGladiator},
	}
	for _, tt := range tests {
		if got := MapRatedTierToTier(tt.rated); got != tt.want {
			t.Errorf("MapRatedTierToTier(%q) = %q, want %q", tt.rated, got, tt.want)
		}
	}
}

func TestArenaStats_String(t *testing.T) {
	tests := []struct {
		name  string
		stats ArenaStats
		want  string
	}{
		{"unranked", ArenaStats{}, "Unranked"},
		{"gold", ArenaStats{Tier: ArenaTierGold, RatedRating: 1234}, "Gold • Rating: 1234"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stats.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
