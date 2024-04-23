import urllib3

from league_rpc.utils.polling import wait_until_exists

urllib3.disable_warnings()


def get_current_ingame_time(default_time: int) -> int:
    """
    Gets the current time of the game.
    """
    url = "https://127.0.0.1:2999/liveclientdata/gamestats"
    if response := wait_until_exists(
        url=url,
        custom_message="""
        Was unable to find the game time.
        Fallback (the time from which you executed this script) is now set as the 'elapsed time' of the game
        "Contact @haze.dev on discord, or submit a ticket on Github.
        """,
    ):
        return int(response.json()["gameTime"])
    return default_time
