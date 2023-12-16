import urllib3

from league_rpc_linux.polling import wait_until_exists

urllib3.disable_warnings()


def summoner_name(with_discriminator: bool = False) -> str:
    """
    Gets the current summoner name.

    if with_discriminator is True, the function will return a summoners name with #EUW / #EUNE etc
        Defaults to not include it.

    """
    url = "https://127.0.0.1:2999/liveclientdata/activeplayername"
    if response := wait_until_exists(
        url=url,
        custom_message="""
            Summoner name could not be found.
            Contact @haze.dev on discord, or submit a ticket on Github.
            """,
    ):
        name = str(response.json())
        return name if with_discriminator else name.split("#", maxsplit=1)[0]

    return ""
