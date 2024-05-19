import urllib3
from requests import Response

from league_rpc.username import get_riot_id
from league_rpc.utils.const import ACTIVE_PLAYER_URL, PLAYER_KDA_SCORES_URL
from league_rpc.utils.polling import wait_until_exists

urllib3.disable_warnings()


def get_kda() -> str:
    """
    Get the current KDA of your game.
    """
    response = get_current_user_stats()
    if isinstance(response, Response):
        parsed_data = response.json()
        kills = str(parsed_data["kills"])
        deaths = str(parsed_data["deaths"])
        assists = str(parsed_data["assists"])

        return f"{kills}/{deaths}/{assists}"
    return ""


def get_level() -> int:
    """
    Get the current Level of your game.
    """
    response = get_current_active_player_stats()
    if isinstance(response, Response):
        parsed_data = response.json()
        level = int(parsed_data["level"])

        return level
    return 0


def get_gold() -> int:
    """
    Get the current gold of your game.
    """
    response = get_current_active_player_stats()
    if isinstance(response, Response):
        parsed_data = response.json()
        gold = int(parsed_data["currentGold"])

        return gold
    return 0


def get_creepscore() -> str:
    """
    Get the current creepScore of your live game
    creepScore is updated every 10cs by Riot.
    """
    response = get_current_user_stats()
    if isinstance(response, Response):
        parsed_data = response.json()
        creep_score = str(parsed_data["creepScore"])
        return f"{creep_score}cs"

    return ""


def get_current_user_stats() -> Response | None:
    """
    Request data from playerscores?riotId and return the response.
    """
    your_riot_id = get_riot_id()
    if your_riot_id:
        # If the summoner name is not found, we don't want the KDA.
        player_score_url = PLAYER_KDA_SCORES_URL.format_map({"riotId": your_riot_id})
        if response := wait_until_exists(url=player_score_url):
            return response
    return None


def get_current_active_player_stats() -> Response | None:
    """
    Request data from liveclientdata/activeplayer and return the response.
    """
    if response := wait_until_exists(url=ACTIVE_PLAYER_URL):
        return response
    return None
