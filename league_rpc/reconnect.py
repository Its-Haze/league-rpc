"""
Holds functions to try to reconnect to discord whenever the pipe closes or discord suddenly crashes.
"""

import sys
import time

import pypresence

from league_rpc.utils.color import Color


def discord_reconnect_attempt(
    rpc: pypresence.Presence,
    amount_of_tries: int = 12,
    amount_of_waiting: int = 5,
):
    """
    Attempts to connect to discord, over a period of time. If it still fails, it will exit the program.
    """
    for i in range(amount_of_tries):
        try:
            time.sleep(amount_of_waiting)
            print(
                f"{Color.yellow}({i + 1}/{amount_of_tries}). Attempting to reconnect..{Color.reset}"
            )
            rpc.connect()
            print(
                f"{Color.green}Successfully reconnected.. Proceeding as normal.{Color.reset}"
            )
            break

        except (
            pypresence.exceptions.DiscordNotFound,
            pypresence.exceptions.DiscordError,
            pypresence.exceptions.InvalidPipe,
            pypresence.exceptions.PipeClosed,
            ConnectionError,
        ):
            pass
    else:
        print(
            f"{Color.red}Was unable to reconnect to Discord. after trying for {amount_of_tries * amount_of_waiting} seconds.{Color.reset}"
        )
        sys.exit()
