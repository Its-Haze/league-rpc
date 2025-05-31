from league_rpc.__version__ import __version__

# Discord Application: League of Legends
DEFAULT_CLIENT_ID = "1194034071588851783"
DISCORD_PROCESS_NAMES: list[str] = [
    "Discord",
    "DiscordPTB",
    "DiscordCanary",
    "electron",
]
LEAGUE_OF_LEGENDS_LOGO = "https://github.com/Its-Haze/league-rpc/blob/master/assets/leagueoflegends.png?raw=true"
SMALL_TEXT = f"its-haze/league-rpc @Github.com ({__version__})"

ALL_GAME_DATA_URL = "https://127.0.0.1:2999/liveclientdata/allgamedata"

ACTIVE_PLAYER_URL = "https://127.0.0.1:2999/liveclientdata/activeplayer"

PLAYER_KDA_SCORES_URL = (
    "https://127.0.0.1:2999/liveclientdata/playerscores?riotId={riotId}"
)

BASE_SKIN_URL = "https://ddragon.leagueoflegends.com/cdn/img/champion/tiles/"

BASE_MAP_ICON_URL = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/content/src/leagueclient/gamemodeassets/{map_name}/img/game-select-icon-active.png"


TFT_COMPANIONS_URL = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/assets"


PROFILE_ICON_BASE_URL = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/v1/profile-icons/{icon_id}.jpg"

DDRAGON_CHAMPION_DATA = "https://ddragon.leagueoflegends.com/cdn/{version}/data/{locale}/champion/{name}.json"

MERAKIANALYTICS_CHAMPION_DATA = (
    "https://cdn.merakianalytics.com/riot/lol/resources/latest/{locale}/champions.json"
)

DDRAGON_API_VERSIONS = "https://ddragon.leagueoflegends.com/api/versions.json"

MAP_ICON_CONVERT_MAP: dict[int, str] = {
    11: "classic_sru",
    12: "aram",
    21: "gamemodex",
    22: "tft",
    30: "cherry",
    33: "strawberry",
    35: "brawl",
}


GAME_MODE_CONVERT_MAP: dict[str, str] = {
    "PRACTICETOOL": "Summoner's Rift (Custom)",
    "ARAM": "Howling Abyss (ARAM)",
    "CLASSIC": "Summoner's Rift",
    "TUTORIAL": "Summoner's Rift (Tutorial)",
    "URF": "Summoner's Rift (URF)",
    "NEXUSBLITZ": "Nexux Blitz",
    "CHERRY": "Arena",
    "STRAWBERRY": "Swarm",
    "BRAWL": "Brawl",
    "TUTORIAL_MODULE_3": "Summoner's Rift (Tutorial)",
    "TUTORIAL_MODULE_2": "Summoner's Rift (Tutorial)",
    "TUTORIAL_MODULE_1": "Summoner's Rift (Tutorial)",
    "ULTBOOK": "Ultimate Spellbook",
    "SWIFTPLAY": "Swiftplay",
}

CHAMPION_NAME_CONVERT_MAP = {
    "AurelionSol": "Aurelion Sol",
    "Belveth": "Bel'Veth",
    "Chogath": "Cho'Gath",
    "DrMundo": "Dr. Mundo",
    "JarvanIV": "Jarvan IV",
    "Kaisa": "Kai'Sa",
    "Khazix": "Kha'Zix",
    "KogMaw": "Kog'Maw",
    "KSante": "K'Sante",
    "LeeSin": "Lee Sin",
    "MasterYi": "Master Yi",
    "MissFortune": "Miss Fortune",
    "Nunu": "Nunu & Willump",
    "RekSai": "Rek'Sai",
    "Renata": "Renata Glasc",
    "TahmKench": "Tahm Kench",
    "TwistedFate": "Twisted Fate",
    "Velkoz": "Vel'Koz",
    "MonkeyKing": "Wukong",
    "XinZhao": "Xin Zhao",
}


LEAGUE_RANKED_EMBLEM = "https://github.com/Its-Haze/league-assets/blob/master/ranked_emblems/{tier}.png?raw=true"
LEAGUE_CHERRY_RANKED_EMBLEM = "https://github.com/Its-Haze/league-assets/blob/master/cherry_rated_medallions/{tier}.png?raw=true"


RANKED_TYPE_MAPPER = {
    "RANKED_SOLO_5x5": "Ranked Solo/Duo",
    "RANKED_FLEX_SR": "Ranked Flex",
    "RANKED_TFT": "Teamfight Tactics (Ranked)",
    "RANKED_TFT_DOUBLE_UP": "Teamfight Tactics (Double Up Workshop)",
    "RANKED_TFT_TURBO": "Teamfight Tactics (Hyper Roll)",
    "CHERRY": "Arena",
    "NORMAL": "Normal (Draft Pick)",
}


# queue_detailed_description if it exists.. otherwise use queue_name
QUEUE_MAPPER = {
    "Normal (Quickplay)",  # detailed description
    "Normal (Draft Pick)",  # # detailed description
    "Ranked Solo/Duo",  # name
    "Ranked Flex",  # name
    "ARAM",  # name
    "Arena",  # name
    "Teamfight Tactics (Normal)",  # detailed description
    "Teamfight Tactics (Ranked)",  # detailed description
    "Teamfight Tactics (Double Up Workshop)",  # detailed description
    "Teamfight Tactics (Hyper Roll)",  # detailed description
}

COMMON_DRIVES = ["C:", "D:", "E:", "F:", "G:", "H:", "I:", "J:", "K:", "L:", "Z:"]
DEFAULT_LEAGUE_CLIENT_EXE_PATH = "\\Riot Games\\Riot Client\\RiotClientServices.exe"
DEFAULT_LEAGUE_CLIENT_EXECUTABLE = "RiotClientServices.exe"


# example: Ahri_86
ANIMATED_SKIN_URL = "https://github.com/Its-Haze/league-assets/blob/master/animated_skins/{filename}.gif?raw=true"

# https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/assets/characters/
# Search for animatedsplash
ANIMATED_SKINS = [
    "Ahri_86",
    "Ezreal_5",
    "Lux_7",
    "MissFortune_16",
    "Samira_30",
    "Seraphine_1",
    "Seraphine_2",
    "Seraphine_3",
    "Sona_6",
    "Udyr_3",
]
