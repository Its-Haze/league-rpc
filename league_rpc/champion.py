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
    ANIMATED_SKIN_URL,
    ANIMATED_SKINS,
    BASE_SKIN_URL,
    CHAMPION_NAME_CONVERT_MAP,
    DDRAGON_CHAMPION_DATA,
    GAME_MODE_CONVERT_MAP,
    MERAKIANALYTICS_CHAMPION_DATA,
)
from league_rpc.utils.polling import wait_until_exists

urllib3.disable_warnings()


def get_specific_champion_data(name: str, locale: str) -> dict[str, Any]:
    """
    Get the specific champion data for the champion name.
    """
    version: str = get_latest_version()

    url = DDRAGON_CHAMPION_DATA.format_map(
        {
            "version": version,
            "name": name,
            "locale": locale,
        }
    )

    try:
        response: requests.Response = requests.get(
            url=url,
            timeout=15,
        )
        response.raise_for_status()
        return response.json()
    except requests.RequestException as e:
        raise RuntimeError(
            f"Failed to fetch champion data for {name} (locale: {locale}, version: {version}). URL: {url}. Error: {e}"
        ) from e
    except ValueError as e:
        raise RuntimeError(
            f"Failed to parse JSON response for champion {name}. URL: {url}. Error: {e}"
        ) from e


def get_specific_chroma_data(name: str, locale: str) -> dict[str, Any]:
    """
    Get the specific chroma champion data for the champion name.
    """
    url = MERAKIANALYTICS_CHAMPION_DATA.format_map(
        {
            "locale": locale.replace("_", "-"),
        }
    )
    try:
        response: requests.Response = requests.get(
            url=url,
            timeout=15,
        )
        response.raise_for_status()
        data = response.json()
        return data[name]
    except requests.RequestException as e:
        raise RuntimeError(
            f"Failed to fetch chroma data for {name} (locale: {locale}). URL: {url}. Error: {e}"
        ) from e
    except (ValueError, KeyError) as e:
        raise RuntimeError(
            f"Failed to parse chroma data for {name}. URL: {url}. Error: {e}"
        ) from e


def gather_game_mode() -> str:
    """
    Get the current game mode.
    """
    if response := wait_until_exists(
        url=ALL_GAME_DATA_URL,
        custom_message="Did not find game data.. Will try again in 5 seconds",
    ):
        parsed_data = response.json()
        game_mode = GAME_MODE_CONVERT_MAP.get(
            parsed_data["gameData"]["gameMode"],
            parsed_data["gameData"]["gameMode"],
        )
        return game_mode
    return ""


def gather_ingame_information(
    silent: bool = False,
) -> tuple[str, str, str, int, str, int, int]:
    """
    Get the current playing champion name.
    """
    your_summoner_name: str = get_riot_id()

    champion_name: str | None = None
    skin_id: int | None = None
    skin_name: str | None = None
    chroma_name: str | None = None
    game_mode: str | None = (
        None  # Set if the game mode was never found.. Maybe you are playing something new?
    )
    level: int | None = None
    gold: int | None = None

    if response := wait_until_exists(
        url=ALL_GAME_DATA_URL,
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

            champion_name, skin_id, skin_name, chroma_name = gather_league_data(
                parsed_data=parsed_data, summoners_name=your_summoner_name
            )
            if game_mode in ("Arena", "Swarm"):
                level, gold = get_level(), get_gold()

            if not silent:
                print("-" * 50)
                if champion_name:
                    print(
                        f"{Color.yellow}Champion name found {Color.green}({CHAMPION_NAME_CONVERT_MAP.get(champion_name, champion_name)}),{Color.yellow} continuing..{Color.reset}"
                    )
                if skin_name:
                    print(
                        f"{Color.yellow}Skin detected: {Color.green}{skin_name},{Color.yellow} continuing..{Color.reset}"
                    )
                if chroma_name:
                    print(
                        f"{Color.yellow}Chroma detected: {Color.green}{chroma_name},{Color.yellow} continuing..{Color.reset}"
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
        (chroma_name or ""),
        (skin_id or 0),
        (game_mode or ""),
        (level or 0),
        (gold or 0),
    )


def fetch_current_player_data(
    all_game_data: list[dict[str, Any]], summoner_name: str
) -> dict[str, Any]:
    """
    Fetch the current player data from the allPlayers list.
    """
    for player in all_game_data:
        if player["riotId"] == summoner_name:
            return player

    raise ValueError(
        f"Unable to find in game data about the summoner ({summoner_name})"
    )


def get_champion_name_from_raw_champion_name(raw_champion_name: str) -> str:
    """
    Extract champion ID from rawChampionName field.
    Format: 'game_character_displayname_ChampionId' -> 'ChampionId'
    """
    return raw_champion_name.split("_")[-1]


def gather_league_data(
    parsed_data: dict[str, Any],
    summoners_name: str,
) -> tuple[Optional[str], int, Optional[str], Optional[str]]:
    """
    If the gamemode is LEAGUE, gather the relevant information and return it to RPC.
    """
    champion_name: Optional[str] = None
    skin_id: int = 0
    base_skin_id: int = 0
    skin_name: Optional[str] = None
    chroma_name: Optional[str] = None
    locale = find_game_locale(
        league_processes=["LeagueClient.exe", "LeagueClientUx.exe"]
    )

    current_summoner_data = fetch_current_player_data(
        all_game_data=parsed_data["allPlayers"],
        summoner_name=summoners_name,
    )

    raw_champion_name = get_champion_name_from_raw_champion_name(
        raw_champion_name=current_summoner_data["rawChampionName"]
    )

    ddragon_champion_data = get_specific_champion_data(
        name=raw_champion_name,
        locale=locale,
    )

    champion_name = ddragon_champion_data["data"][raw_champion_name]["id"]
    skin_id = current_summoner_data.get("skinID", 0)

    base_skin_id = get_base_skin_id_from_skin_id(
        skin_id, raw_champion_name, ddragon_champion_data
    )
    skin_name = get_skin_name_from_id(
        base_skin_id, raw_champion_name, ddragon_champion_data
    )

    if skin_is_chroma(skin_id, base_skin_id):
        chroma_data = get_specific_chroma_data(
            name=raw_champion_name,
            locale="en-US",
        )
        chroma_name = get_chroma_name(skin_id, base_skin_id, chroma_data)

    return champion_name, base_skin_id, skin_name, chroma_name


def get_chroma_name(
    skin_id: int, base_skin_id: int, chroma_data: dict[str, Any]
) -> str:
    """
    Get the chroma name for the skin
    
    Some chromas do not exist in the MerakiAnalytics api, so return an empty string in that case.
    """
    # Find the skin data matching the base_skin_id
    matching_skins = [
        x for x in chroma_data["skins"] if str(x["id"]).endswith(str(base_skin_id))
    ]

    if not matching_skins:
        return ""

    _skin_data: dict[str, Any] = matching_skins[0]

    # Find the chroma name matching the skin_id
    matching_chromas = [
        x["name"] for x in _skin_data.get("chromas", []) if str(x["id"]).endswith(str(skin_id))
    ]

    if not matching_chromas:
        return ""

    return matching_chromas[0]


def skin_is_chroma(skin_id: int, base_skin_id: int) -> bool:
    """Check if the skin is a chroma"""
    return skin_id != base_skin_id


def get_base_skin_id_from_skin_id(
    skin_id: int, raw_champion_name: str, ddragon_champion_data: dict[str, Any]
) -> int:
    """
    Determine the base skin ID from the API's skinID.

    If skinID matches a skin's 'num' in DDragon, it's a base skin.
    If not, it's likely a chroma - use heuristic to find the parent skin.
    """
    skins = ddragon_champion_data["data"][raw_champion_name]["skins"]

    # Check if skinID is a base skin
    for skin in skins:
        if skin["num"] == skin_id:
            return skin_id

    # Likely a chroma - find the base skin by finding highest num <= skinID
    # This heuristic works because chroma IDs are typically assigned after their base skin
    for skin in sorted(skins, key=lambda x: x["num"], reverse=True):
        if skin["num"] <= skin_id:
            return skin["num"]

    return 0  # Fallback to default skin


def get_skin_name_from_id(
    skin_id: int, raw_champion_name: str, ddragon_champion_data: dict[str, Any]
) -> Optional[str]:
    """Get the skin name from DDragon. Returns None for default skin or if not found."""
    skins = ddragon_champion_data["data"][raw_champion_name]["skins"]

    for skin in skins:
        if skin["num"] == skin_id:
            return None if skin["name"] == "default" else skin["name"]

    return None


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

        # If the Champ_skinID matches a animated skin, then return the URL for the animated skin instead.
        if f"{champion_name}_{skin_id}" in ANIMATED_SKINS:
            url = ANIMATED_SKIN_URL.format_map(
                {"filename": f"{champion_name}_{skin_id}"}
            )
        return url
    url = f"{BASE_SKIN_URL}{champion_name}_0.jpg"
    return url


def check_url(url: str) -> bool:
    """
    Sends a HEAD request to the URL and,
    returns a boolean value depending on if the request,
    was successful (200 OK) or not.
    """
    return requests.head(url=url, timeout=15).status_code == HTTPStatus.OK
