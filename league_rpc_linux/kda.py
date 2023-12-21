import urllib3
from requests import Response

from league_rpc_linux.polling import wait_until_exists
from league_rpc_linux.username import get_summoner_name

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
    Request data from playerscores?summonerName and return the response.
    """
    your_summoner_name = get_summoner_name()
    if your_summoner_name:
        # If the summoner name is not found, we don't want the KDA.

        player_score_url = f"https://127.0.0.1:2999/liveclientdata/playerscores?summonerName={your_summoner_name}"
        if response := wait_until_exists(url=player_score_url):
            return response
    return None
