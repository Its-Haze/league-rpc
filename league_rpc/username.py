from typing import Any

import urllib3

from league_rpc.utils.const import ACTIVE_PLAYER_URL
from league_rpc.utils.polling import wait_until_exists

urllib3.disable_warnings()


def get_riot_id(without_discriminator: bool = False) -> str:
    """
    Gets the current summoner name.

    if without_discriminator is True, the function will not return a summoners name with #EUW / #EUNE etc
        Defaults to include it.

    """
    if response := wait_until_exists(
        url=ACTIVE_PLAYER_URL,
        custom_message="""
            Summoner name could not be found.
            Contact @haze.dev on discord, or submit a ticket on Github.
            """,
    ):
        _response: dict[str, Any] = response.json()
        name_without_discriminator = _response["riotIdGameName"]
        riot_id = _response["riotId"]

        return name_without_discriminator if without_discriminator else riot_id

    return ""
