package discord

import (
	"fmt"
	"math/rand"

	"github.com/its-haze/league-rpc/pkg/types"
)

const (
	// Base URLs for assets
	communityDragonBaseURL = "https://raw.communitydragon.org/latest"
	datadragonBaseURL      = "https://ddragon.leagueoflegends.com/cdn"
	githubAssetsBaseURL    = "https://github.com/Its-Haze/league-assets/blob/master"
	leagueLogoURL          = "https://github.com/Its-Haze/league-rpc/blob/master/assets/league-classic-borderless.jpg?raw=true"
	leagueLogoLargeURL     = "https://github.com/Its-Haze/league-rpc/blob/master/assets/leagueoflegends.png?raw=true"
	tftCompanionsBaseURL   = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/assets"
)

// Animated Ultimate skins (use .gif from GitHub)
var animatedSkins = map[string]string{
	"Ahri_86":        "Ahri_86.gif",        // Immortalized Legend Ahri
	"Ezreal_5":       "Ezreal_5.gif",       // Pulsefire Ezreal
	"Jinx_60":        "Jinx_60.gif",        // Arcane Fractured Jinx
	"Kaisa_71":       "Kaisa_71.gif",       // Immortalized Legend Kai'Sa
	"Lux_7":          "Lux_7.gif",          // Elementalist Lux
	"MissFortune_16": "MissFortune_16.gif", // Gun Goddess Miss Fortune
	"Mordekaiser_54": "Mordekaiser_54.gif", // Sahn-Uzal Mordekaiser
	"Morgana_80":     "Morgana_80.gif",     // Spirit Blossom Morgana
	"Samira_30":      "Samira_30.gif",      // Soul Fighter Samira
	"Seraphine_1":    "Seraphine_1.gif",    // K/DA ALL OUT Seraphine
	"Seraphine_2":    "Seraphine_2.gif",    // Indie Seraphine
	"Seraphine_3":    "Seraphine_3.gif",    // Graceful Phoenix Seraphine
	"Sett_66":        "Sett_66.gif",        // Radiant Serpent Sett
	"Sona_6":         "Sona_6.gif",         // DJ Sona
	"Udyr_3":         "Udyr_3.gif",         // Spirit Guard Udyr
}

// launchingPlaceholderSkins is the pool the "launching" placeholder's large
// image is randomly picked from on every rotation, same as the reference.
var launchingPlaceholderSkins = []string{
	"Ahri_86.gif",
	"Ezreal_5.gif",
	"Lux_7.gif",
	"MissFortune_16.gif",
	"Samira_30.gif",
	"Seraphine_1.gif",
	"Seraphine_2.gif",
	"Seraphine_3.gif",
	"Sona_6.gif",
	"Udyr_3.gif",
}

// GetLaunchingPlaceholderImageURL picks a random skin from
// launchingPlaceholderSkins, matching the reference's random.choice.
func GetLaunchingPlaceholderImageURL() string {
	filename := launchingPlaceholderSkins[rand.Intn(len(launchingPlaceholderSkins))]
	return fmt.Sprintf("%s/animated_skins/%s?raw=true", githubAssetsBaseURL, filename)
}

// GetProfileIconURL returns the URL for a summoner profile icon
func GetProfileIconURL(iconID int) string {
	return fmt.Sprintf("%s/plugins/rcp-be-lol-game-data/global/default/v1/profile-icons/%d.jpg",
		communityDragonBaseURL, iconID)
}

// GetRankEmblemURL returns the URL for a ranked emblem
func GetRankEmblemURL(tier types.Tier) string {
	if tier == "" {
		return ""
	}
	return fmt.Sprintf("%s/ranked_emblems/%s.png?raw=true", githubAssetsBaseURL, tier.DisplayName())
}

// GetArenaEmblemURL returns the URL for an Arena medallion
func GetArenaEmblemURL(tier types.ArenaTier) string {
	if tier == "" {
		return ""
	}
	return fmt.Sprintf("%s/cherry_rated_medallions/%s.png?raw=true", githubAssetsBaseURL, tier)
}

// GetChampionSkinURL returns the URL for a champion skin tile
// Handles animated skins (Ultimate skins) automatically
func GetChampionSkinURL(championName string, skinID int) string {
	// Check if this is an animated skin
	skinKey := fmt.Sprintf("%s_%d", championName, skinID)
	if animatedFilename, isAnimated := animatedSkins[skinKey]; isAnimated {
		return fmt.Sprintf("%s/animated_skins/%s?raw=true", githubAssetsBaseURL, animatedFilename)
	}

	// Standard skin tile from DDragon
	return fmt.Sprintf("%s/img/champion/tiles/%s_%d.jpg", datadragonBaseURL, championName, skinID)
}

// GetMapIconURL returns the URL for a map icon based on map ID
func GetMapIconURL(mapID types.MapID) string {
	var mapName string

	switch mapID {
	case types.MapSummonersRift:
		mapName = "classic_sru"
	case types.MapHowlingAbyss:
		mapName = "aram"
	case types.MapTFT:
		mapName = "tft"
	case types.MapArena:
		mapName = "cherry"
	case types.MapSwarm:
		mapName = "strawberry"
	case types.MapBrawl:
		mapName = "brawl"
	default:
		mapName = "classic_sru" // Fallback to Summoner's Rift
	}

	return fmt.Sprintf("%s/plugins/rcp-be-lol-game-data/global/default/content/src/leagueclient/gamemodeassets/%s/img/game-select-icon-hover.png",
		communityDragonBaseURL, mapName)
}

// GetLeagueLogoURL returns the URL for the League of Legends logo
func GetLeagueLogoURL() string {
	return leagueLogoURL
}

// GetLeagueLogoLargeURL returns the square League of Legends logo used as
// the large image when spectating a game whose champion can't be resolved.
func GetLeagueLogoLargeURL() string {
	return leagueLogoLargeURL
}

// GetTFTCompanionURL builds the URL for a TFT companion icon from the
// lowercased path segment after "ASSETS/" in the LCU's loadoutsIcon field.
func GetTFTCompanionURL(iconPath string) string {
	if iconPath == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", tftCompanionsBaseURL, iconPath)
}

// gameModeDisplayNames maps raw LCU gameMode values to display names.
// "Nexux Blitz" is intentional, not a typo introduced here.
var gameModeDisplayNames = map[types.GameMode]string{
	"PRACTICETOOL":      "Summoner's Rift (Custom)",
	"ARAM":              "Howling Abyss (ARAM)",
	"CLASSIC":           "Summoner's Rift",
	"TUTORIAL":          "Summoner's Rift (Tutorial)",
	"URF":               "Summoner's Rift (URF)",
	"NEXUSBLITZ":        "Nexux Blitz",
	"CHERRY":            "Arena",
	"STRAWBERRY":        "Swarm",
	"BRAWL":             "Brawl",
	"KIWI":              "ARAM: Mayhem",
	"KIWI_JADE":         "ARAM: Mayhem Classic-ish",
	"JADE":              "League Classic",
	"TUTORIAL_MODULE_3": "Summoner's Rift (Tutorial)",
	"TUTORIAL_MODULE_2": "Summoner's Rift (Tutorial)",
	"TUTORIAL_MODULE_1": "Summoner's Rift (Tutorial)",
	"ULTBOOK":           "Ultimate Spellbook",
	"SWIFTPLAY":         "Swiftplay",
	"RUBY":              "Doom Bots",
	"RUBY_TRIAL_1":      "Doom Bots - Veigar's Evil!",
	"RUBY_TRIAL_3":      "Doom Bots - Veigar's Doom!",
}

// FormatGameModeName converts a raw LCU gameMode value to its display name,
// falling back to the raw value itself for anything not in the map.
func FormatGameModeName(gameMode types.GameMode) string {
	if name, ok := gameModeDisplayNames[gameMode]; ok {
		return name
	}
	return string(gameMode)
}

// FormatKDA formats KDA and CS into a display string
func FormatKDA(kills, deaths, assists, cs int) string {
	return fmt.Sprintf("%d/%d/%d · %dcs", kills, deaths, assists, cs)
}

// FormatArenaStats formats the Arena in-game stat line: KDA, level, gold.
func FormatArenaStats(kills, deaths, assists, level, gold int) string {
	return fmt.Sprintf("%d/%d/%d · lvl: %d · gold: %d", kills, deaths, assists, level, gold)
}

// FormatSwarmStats formats the Swarm in-game stat line: CS, level, gold.
func FormatSwarmStats(cs, level, gold int) string {
	return fmt.Sprintf("%dcs · lvl: %d · gold: %d", cs, level, gold)
}

// FormatSkinName formats the skin name for display, since most skin names
func FormatSkinName(championName, skinName, chromaName string) string {
	if chromaName != "" {
		return fmt.Sprintf("%s (%s)", skinName, chromaName)
	}
	if skinName != "" && skinName != "default" {
		return skinName
	}
	return championName
}
