from typing import Any, Optional

import requests
import urllib3

from league_rpc_linux.colors import Colors
from league_rpc_linux.polling import wait_until_exists
from league_rpc_linux.username import get_summoner_name

urllib3.disable_warnings()


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
}


def gather_ingame_information() -> tuple[str, int, str, int]:
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

    if response := wait_until_exists(
        url=all_game_data_url,
        custom_message="Did not find game data.. Will try again in 5 seconds",
    ):
        parsed_data = response.json()
        game_mode = game_mode_convert_map.get(
            parsed_data["gameData"]["gameMode"],
            parsed_data["gameData"]["gameMode"],
        )
        if game_mode != "TFT":
            # If the gamemode is LEAGUE gather the relevant information.
            champion_name, skin_id, skin_name = gather_league_data(
                parsed_data=parsed_data, summoners_name=your_summoner_name
            )
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
        else:
            # If the currentGame is TFT.. gather the relevant information
            level = gather_tft_data(parsed_data=parsed_data)

    # Returns default values if information was not found.
    return (
        (champion_name or "???"),
        (skin_id or 0),
        (game_mode or "???"),
        (level or 0),
    )


def gather_league_data(
    parsed_data: dict[str, Any], summoners_name: str
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


def gather_tft_data(parsed_data: dict[str, Any]) -> Optional[int]:
    """
    If the gamemode is TFT, it will gather information and return it to RPC
    """
    level: Optional[int] = None

    level = int(parsed_data["activePlayer"]["level"])
    return level


def get_skin_asset(champion_name: str, skin_id: int) -> str:
    """
    Returns either a default champion art
    or the selected skin for that specific champion.
    """

    if skin_id != 0:
        # Return the skin art
        url = f"https://ddragon.leagueoflegends.com/cdn/img/champion/tiles/{champion_name}_{skin_id}.jpg"
        try:
            if requests.get(url=url, verify=False, timeout=15).status_code == 200:
                print(f"{Colors.green}Successfully found skin art.{Colors.reset}")
        except requests.RequestException as exc:
            print(
                f"{Colors.red}Caught exception while trying to find skin image. {exc}\nThe URL was {url}{Colors.reset}"
            )
        return url

    # Otherwise return the default champion art.
    url = f"http://ddragon.leagueoflegends.com/cdn/13.10.1/img/champion/{champion_name}.png"
    try:
        if requests.get(url=url, verify=False, timeout=15).status_code == 200:
            print(
                f"{Colors.green}Default skin detected on {champion_name}{Colors.reset}"
            )
    except requests.RequestException as exc:
        print(
            f"{Colors.red}Caught exception while trying to find default skin for {champion_name}. {exc}\nThe URL was {url}{Colors.reset}"
        )
    return url
