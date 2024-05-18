from http import HTTPStatus
from typing import Any, Optional

import requests
import urllib3

from league_rpc.disable_native_rpc.disable import find_game_locale
from league_rpc.kda import get_gold, get_level
from league_rpc.latest_version import get_latest_version
from league_rpc.username import get_riot_id
from league_rpc.utils.color import Color
from league_rpc.utils.const import (
    ALL_GAME_DATA_URL,
    BASE_SKIN_URL,
    CHAMPION_NAME_CONVERT_MAP,
    DDRAGON_CHAMPION_DATA,
    GAME_MODE_CONVERT_MAP,
)
from league_rpc.utils.polling import wait_until_exists

urllib3.disable_warnings()


def get_specific_champion_data(name: str, locale: str) -> dict[str, Any]:
    """
    Get the specific champion data for the champion name.
    """
    response: requests.Response = requests.get(
        url=DDRAGON_CHAMPION_DATA.format_map(
            {
                "version": get_latest_version(),
                "name": name,
                "locale": locale,
            }
        ),
        timeout=15,
    )
    return response.json()


def gather_ingame_information() -> tuple[str, str, int, str, int, int]:
    """
    Get the current playing champion name.
    """
    all_game_data_url = ALL_GAME_DATA_URL
    your_summoner_name: str = get_riot_id()

    champion_name: str | None = None
    skin_id: int | None = None
    skin_name: str | None = None
    game_mode: str | None = (
        None  # Set if the game mode was never found.. Maybe you are playing something new?
    )
    level: int | None = None
    gold: int | None = None

    if response := wait_until_exists(
        url=all_game_data_url,
        custom_message="Did not find game data.. Will try again in 5 seconds",
    ):
        parsed_data = response.json()
        game_mode = GAME_MODE_CONVERT_MAP.get(
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
                    f"{Color.yellow}Champion name found {Color.green}({CHAMPION_NAME_CONVERT_MAP.get(champion_name, champion_name)}),{Color.yellow} continuing..{Color.reset}"
                )
            if skin_name:
                print(
                    f"{Color.yellow}Skin detected: {Color.green}{skin_name},{Color.yellow} continuing..{Color.reset}"
                )
            if game_mode:
                print(
                    f"{Color.yellow}Game mode detected: {Color.green}{game_mode},{Color.yellow} continuing..{Color.reset}"
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
) -> tuple[Optional[str], int, Optional[str]]:
    """
    If the gamemode is LEAGUE, gather the relevant information and return it to RPC.
    """
    champion_name: Optional[str] = None
    skin_id: int = 0
    skin_name: Optional[str] = None

    for player in parsed_data["allPlayers"]:
        if player["riotId"] == summoners_name:
            raw_champion_name: str = player["rawChampionName"].split("_")[-1]
            champion_data: dict[str, Any] = get_specific_champion_data(
                name=raw_champion_name,
                locale=find_game_locale(
                    league_processes=[
                        "LeagueClient.exe",
                        "LeagueClientUx.exe",
                    ]
                ),
            )
            champion_name = champion_data["data"][raw_champion_name]["id"]
            skin_name = player.get("skinName", None)

            if skin_name:
                skin_id = [
                    x["num"]
                    for x in champion_data["data"][raw_champion_name]["skins"]
                    if x["name"] == skin_name
                ][0]

            break
        continue
    return champion_name, skin_id, skin_name


def get_skin_asset(
    champion_name: str,
    skin_id: int,
) -> str:
    """
    Returns the URL for the skin/default skin of the champion.
    If a chroma has been selected, it will return the base skin for that chroma.
        Since RIOT does not have individual images for each chroma.
    """

    while skin_id:
        url: str = f"{BASE_SKIN_URL}{champion_name}_{skin_id}.jpg"
        if not check_url(url=url):
            skin_id -= 1
            continue

        print("in while: ", url)
        return url

    url = f"{BASE_SKIN_URL}{champion_name}_0.jpg"

    print("outside while", url)
    return url


def check_url(url: str) -> bool:
    """
    Sends a HEAD request to the URL and,
    returns a boolean value depending on if the request,
    was successful (200 OK) or not.
    """
    return requests.head(url=url, timeout=15).status_code == HTTPStatus.OK
