from typing import Any, Optional

import urllib3

from league_rpc_linux.colors import Colors
from league_rpc_linux.kda import get_gold, get_level
from league_rpc_linux.polling import wait_until_exists
from league_rpc_linux.username import get_summoner_name

urllib3.disable_warnings()


BASE_SKIN_URL = "https://ddragon.leagueoflegends.com/cdn/img/champion/tiles/"
BASE_CHAMPION_URL = "http://ddragon.leagueoflegends.com/cdn/"


champion_name_convert_map = {
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

game_mode_convert_map = {
    "PRACTICETOOL": "Summoner's Rift (Custom)",
    "ARAM": "Howling Abyss (ARAM)",
    "CLASSIC": "Summoner's Rift",
    "TUTORIAL": "Summoner's Rift (Tutorial)",
    "URF": "Summoner's Rift (URF)",
    "NEXUSBLITZ": "Nexux Blitz",
    "CHERRY": "Arena",
}


def gather_ingame_information() -> tuple[str, str, int, str, int, int]:
    """
    Get the current playing champion name.
    """
    all_game_data_url = "https://127.0.0.1:2999/liveclientdata/allgamedata"
    your_summoner_name = get_summoner_name()

    champion_name: str | None = None
    skin_id: int | None = None
    skin_name: str | None = None
    game_mode: str | None = None  # Set if the game mode was never found.. Maybe you are playing something new?
    level: int | None = None
    gold: int | None = None

    if response := wait_until_exists(
        url=all_game_data_url,
        custom_message="Did not find game data.. Will try again in 5 seconds",
    ):
        parsed_data = response.json()
        game_mode = game_mode_convert_map.get(
            parsed_data["gameData"]["gameMode"],
            parsed_data["gameData"]["gameMode"],
        )

        if game_mode == "TFT":
            # If the currentGame is TFT.. gather the relevant information
            level = get_level()
        else:
            # If the gamemode is LEAGUE gather the relevant information.
            champion_name, skin_id, skin_name = gather_league_data(
                parsed_data=parsed_data, summoners_name=your_summoner_name
            )
            if game_mode == "Arena":
                level, gold = get_level(), get_gold()
            print("-" * 50)
            if champion_name:
                print(
                    f"{Colors.yellow}Champion name found {Colors.green}({champion_name}),{Colors.yellow} continuing..{Colors.reset}"
                )
            if skin_name:
                print(
                    f"{Colors.yellow}Skin detected: {Colors.green}{skin_name},{Colors.yellow} continuing..{Colors.reset}"
                )
            if game_mode:
                print(
                    f"{Colors.yellow}Game mode detected: {Colors.green}{game_mode},{Colors.yellow} continuing..{Colors.reset}"
                )
            print("-" * 50)

    # Returns default values if information was not found.
    return (
        (champion_name or ""),
        (skin_name or ""),
        (skin_id or 0),
        (game_mode or ""),
        (level or 0),
        (gold or 0),
    )


def gather_league_data(
    parsed_data: dict[str, Any],
    summoners_name: str,
) -> tuple[Optional[str], Optional[int], Optional[str]]:
    """
    If the gamemode is LEAGUE, gather the relevant information and return it to RPC.
    """
    champion_name: Optional[str] = None
    skin_id: Optional[int] = None
    skin_name: Optional[str] = None

    for player in parsed_data["allPlayers"]:
        if player["summonerName"] == summoners_name:
            champion_name = champion_name_convert_map.get(
                player["championName"],
                player["championName"],
            )
            skin_id = player["skinID"]
            skin_name = player.get("skinName")
            break
        continue
    return champion_name, skin_id, skin_name


def get_skin_asset(
    champion_name: str,
    skin_id: int,
    patch: str,
    fallback_asset: str,
) -> str:
    """
    Returns either a default champion art
    or the selected skin for that specific champion.
    """
    if skin_id != 0:
        url = f"{BASE_SKIN_URL}{champion_name}_{skin_id}.jpg"
    else:
        url = f"{BASE_CHAMPION_URL}{patch}/img/champion/{champion_name}.png"

    if not wait_until_exists(url):
        print(
            f"""{Colors.red}Failed to request the champion/skin image
    {Colors.orange}Reasons for this could be the following:
{Colors.blue}(1) Maybe a false positive.. A new attempt will be made to find the skin art. But if it keeps failing, then something is wrong.
    If the skin art is after further attempts found, then you can simply ignore this message..
(2) Your version of this application is outdated
(3) The maintainer of this application has not updated to the latest patch..
    If league's latest patch isn't {patch}, then contact ({Colors.orange}@haze.dev{Colors.blue} on Discord).{Colors.reset}"""
        )
        return fallback_asset

    return url
