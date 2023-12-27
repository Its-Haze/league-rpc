# Discord Application: League of Linux
DEFAULT_CLIENT_ID = "1185274747836174377"
DISCORD_PROCESS_NAMES = ["Discord", "DiscordPTB", "DiscordCanary", "electron"]
LEAGUE_OF_LEGENDS_LOGO = "https://github.com/Its-Haze/league-rpc-linux/blob/master/assets/leagueoflegends.png?raw=true"
SMALL_TEXT = "github.com/Its-Haze/league-rpc-linux"
CURRENT_PATCH = "13.24.1"


ALL_GAME_DATA_URL = "https://127.0.0.1:2999/liveclientdata/allgamedata"

BASE_SKIN_URL = "https://ddragon.leagueoflegends.com/cdn/img/champion/tiles/"
BASE_CHAMPION_URL = (
    f"http://ddragon.leagueoflegends.com/cdn/{CURRENT_PATCH}/img/champion/"
)
BASE_MAP_URL = f"http://ddragon.leagueoflegends.com/cdn/{CURRENT_PATCH}/img/map/map"

BASE_MAP_ICON_URL = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/content/src/leagueclient/gamemodeassets/{map_name}/img/game-select-icon-hover.png"

PROFILE_ICON_BASE_URL = (
    f"https://ddragon.leagueoflegends.com/cdn/{CURRENT_PATCH}/img/profileicon/"
)

MAP_ICON_CONVERT_MAP = {
    11: "classic_sru",
    12: "aram",
    21: "gamemodex",
    22: "tft",
    30: "cherry",
}


GAME_MODE_CONVERT_MAP = {
    "PRACTICETOOL": "Summoner's Rift (Custom)",
    "ARAM": "Howling Abyss (ARAM)",
    "CLASSIC": "Summoner's Rift",
    "TUTORIAL": "Summoner's Rift (Tutorial)",
    "URF": "Summoner's Rift (URF)",
    "NEXUSBLITZ": "Nexux Blitz",
    "CHERRY": "Arena",
}

CHAMPION_NAME_CONVERT_MAP = {
    "Aurelion Sol": "AurelionSol",
    "Cho'Gath": "Chogath",
    "Renata Glasc": "Renata",
    "Dr. Mundo": "DrMundo",
    "Miss Fortune": "MissFortune",
    "Kai'Sa": "KaiSa",
    "Kog'Maw": "KogMaw",
    "Rek'Sai": "RekSai",
    "K'Sante": "KSante",
    "Kha'Zix": "KhaZix",
    "Nunu & Willump": "Nunu",
    "Twisted Fate": "TwistedFate",
    "Tahm Kench": "TahmKench",
    "Vel'Koz": "Velkoz",
    "Xin Zhao": "XinZhao",
    "Master Yi": "MasterYi",
}
