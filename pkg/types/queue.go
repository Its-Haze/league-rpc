package types

// QueueID represents different League of Legends queue types
type QueueID int

const (
	QueueSoloQ QueueID = 420
	QueueFlex  QueueID = 440
	QueueARAM  QueueID = 450
	QueueTFT   QueueID = 1090
	QueueArena QueueID = 1700
)

// GameMode represents different game modes
type GameMode string

const (
	GameModeClassic           GameMode = "CLASSIC"
	GameModeARAM              GameMode = "ARAM"
	GameModeTFT               GameMode = "TFT"
	GameModeArena             GameMode = "CHERRY"
	GameModeUltimateSpellbook GameMode = "ULTBOOK"
	GameModeSwarm             GameMode = "STRAWBERRY"
	GameModeBrawl             GameMode = "BRAWL"
)

// MapID represents different League of Legends maps
type MapID int

const (
	MapSummonersRift MapID = 11
	MapHowlingAbyss  MapID = 12
	MapTFT           MapID = 22
	MapArena         MapID = 30
	MapSwarm         MapID = 33
	MapBrawl         MapID = 35
)

// MapName returns the display name for a map ID
func (m MapID) MapName() string {
	switch m {
	case MapSummonersRift:
		return "Summoner's Rift"
	case MapHowlingAbyss:
		return "Howling Abyss"
	case MapTFT:
		return "Convergence"
	case MapArena:
		return "Rings of Wrath"
	case MapSwarm:
		return "Substructure 43"
	case MapBrawl:
		return "Noxian Arena"
	default:
		return "Unknown Map"
	}
}

// Availability represents player online status
type Availability string

const (
	AvailabilityOnline Availability = "Online"
	AvailabilityAway   Availability = "Away"
	AvailabilityDND    Availability = "dnd" // Do Not Disturb
)
