from typing import Any

import urllib3

from league_rpc.utils.polling import wait_until_exists

urllib3.disable_warnings()


def get_riot_id(without_discriminator: bool = False) -> str:
    """
    Gets the current summoner name.

    if without_discriminator is True, the function will not return a summoners name with #EUW / #EUNE etc
        Defaults to include it.

    """

    url = "https://127.0.0.1:2999/liveclientdata/playerlist"
    if response := wait_until_exists(
        url=url,
        custom_message="""
            Summoner name could not be found.
            Contact @haze.dev on discord, or submit a ticket on Github.
            """,
    ):
        _response: list[dict[str, Any]] = response.json()
        name_without_discriminator = _response[0]["riotIdGameName"]
        riot_id = _response[0]["riotId"]

        return name_without_discriminator if without_discriminator else riot_id

    return ""
