import urllib3

from league_rpc_linux.polling import wait_until_exists

urllib3.disable_warnings()


def get_game_mode():
    """
    Gets the current game mode.
    """
    game_mode = "???"  # Set if the game mode was never found.. Maybe you are playing something new?
    url = "https://127.0.0.1:2999/liveclientdata/allgamedata"
    if response := wait_until_exists(
        url=url, custom_message="Game mode could not be found.. Probably not in game.."
    ):
        parsed_data = response.json()
        game_mode = parsed_data["gameData"]["gameMode"]

    match game_mode:
        case "PRACTICETOOL":
            return "Summoner's Rift (Custom)"
        case "ARAM":
            return "Howling Abyss (ARAM)"
        case "CLASSIC":
            return "Summoner's Rift"
        case "TUTORIAL":
            return "Summoner's Rift (Tutorial)"
        case "URF":
            return "Summoner's Rift (URF)"
        case "NEXUSBLITZ":
            return "Nexux Blitz"
        case _:
            return "???"
