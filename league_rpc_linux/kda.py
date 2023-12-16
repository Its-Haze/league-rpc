import urllib3

from league_rpc_linux.polling import wait_until_exists
from league_rpc_linux.username import summoner_name

urllib3.disable_warnings()


def get_kda() -> str:
    """
    Get the current KDA + creepScore of your live game.

    creepScore is updated every 10cs by Riot.
    """
    your_summoner_name = summoner_name()
    if your_summoner_name:
        # If the summoner name is not found, we don't want the KDA.

        player_score_url = f"https://127.0.0.1:2999/liveclientdata/playerscores?summonerName={your_summoner_name}"
        if response := wait_until_exists(url=player_score_url):
            parsed_data = response.json()
            kills = str(parsed_data["kills"])
            deaths = str(parsed_data["deaths"])
            assists = str(parsed_data["assists"])
            creep_score = str(parsed_data["creepScore"])

            return f"· {kills}/{deaths}/{assists} · {creep_score}cs"
    return ""
