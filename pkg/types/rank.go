package types

import (
	"fmt"
	"strings"
)

// Tier represents ranked tiers
type Tier string

const (
	TierIron        Tier = "IRON"
	TierBronze      Tier = "BRONZE"
	TierSilver      Tier = "SILVER"
	TierGold        Tier = "GOLD"
	TierPlatinum    Tier = "PLATINUM"
	TierEmerald     Tier = "EMERALD"
	TierDiamond     Tier = "DIAMOND"
	TierMaster      Tier = "MASTER"
	TierGrandmaster Tier = "GRANDMASTER"
	TierChallenger  Tier = "CHALLENGER"
)

// Division represents ranked divisions (IV, III, II, I)
type Division string

const (
	DivisionIV  Division = "IV"
	DivisionIII Division = "III"
	DivisionII  Division = "II"
	DivisionI   Division = "I"
	DivisionNA  Division = "NA" // For Master+
)

// DisplayName title-cases the tier for display, matching the reference
// implementation's plain .capitalize() (e.g. "GOLD" -> "Gold").
func (t Tier) DisplayName() string {
	s := string(t)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// RankedStats represents ranked statistics for SoloQ/Flex/TFT
type RankedStats struct {
	Tier         Tier     `json:"tier"`
	Division     Division `json:"division"`
	LeaguePoints int      `json:"league_points"`
}

// String returns a formatted string representation of the rank
// Example: "Gold IV: 57 LP"
func (r *RankedStats) String() string {
	if r.Tier == "" {
		return "Unranked"
	}

	divisionPart := ""
	if r.Division != "" && r.Division != DivisionNA {
		divisionPart = " " + string(r.Division)
	}

	return fmt.Sprintf("%s%s: %d LP", r.Tier.DisplayName(), divisionPart, r.LeaguePoints)
}

// IsEmpty returns true if the rank data is empty/unranked
func (r *RankedStats) IsEmpty() bool {
	return r.Tier == ""
}

// ArenaRatedTier represents Arena rated tiers
type ArenaRatedTier string

const (
	ArenaRatedNone   ArenaRatedTier = "NONE"
	ArenaRatedGray   ArenaRatedTier = "GRAY"
	ArenaRatedGreen  ArenaRatedTier = "GREEN"
	ArenaRatedBlue   ArenaRatedTier = "BLUE"
	ArenaRatedPurple ArenaRatedTier = "PURPLE"
	ArenaRatedOrange ArenaRatedTier = "ORANGE"
)

// ArenaTier represents Arena display tiers. Values are already the display
// strings shown in presence text, matching the reference implementation.
type ArenaTier string

const (
	ArenaTierWood      ArenaTier = "Wood"
	ArenaTierBronze    ArenaTier = "Bronze"
	ArenaTierSilver    ArenaTier = "Silver"
	ArenaTierGold      ArenaTier = "Gold"
	ArenaTierGladiator ArenaTier = "Gladiator"
)

// ArenaStats represents Arena ranked statistics
type ArenaStats struct {
	RatedTier   ArenaRatedTier `json:"rated_tier"`
	Tier        ArenaTier      `json:"tier"`
	RatedRating int            `json:"rated_rating"`
}

// MapRatedTierToTier maps Riot's rated tier to display tier
func MapRatedTierToTier(ratedTier ArenaRatedTier) ArenaTier {
	switch ratedTier {
	case ArenaRatedGray:
		return ArenaTierWood
	case ArenaRatedGreen:
		return ArenaTierBronze
	case ArenaRatedBlue:
		return ArenaTierSilver
	case ArenaRatedPurple:
		return ArenaTierGold
	case ArenaRatedOrange:
		return ArenaTierGladiator
	default: // ArenaRatedNone and anything unrecognized
		return ""
	}
}

// String returns a formatted string representation of Arena rank
func (a *ArenaStats) String() string {
	if a.Tier == "" {
		return "Unranked"
	}
	return strings.TrimSpace(fmt.Sprintf("%s • Rating: %d", a.Tier, a.RatedRating))
}

// IsEmpty returns true if Arena rank data is empty
func (a *ArenaStats) IsEmpty() bool {
	return a.Tier == ""
}
