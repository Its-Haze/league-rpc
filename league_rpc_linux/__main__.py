import sys
import time

import pypresence

from league_rpc_linux.champion import champion_asset, get_champion_name
from league_rpc_linux.colors import Colors
from league_rpc_linux.gamemode import get_game_mode
from league_rpc_linux.gametime import get_current_ingame_time
from league_rpc_linux.kda import get_kda
from league_rpc_linux.processes.process import (
    check_discord_process,
    check_league_client_process,
    check_riot_games_service_process,
    player_state,
)


def main(colors: Colors):
    """
    main program.
    """
    # Initialize
    print(colors.logo)

    discord_process_names = ["Discord", "DiscordPTB", "DiscordCanary", "electron"]
    rpc = check_discord_process(colors, discord_process_names)

    check_riot_games_service_process(colors)
    check_league_client_process(colors)

    print(f"{colors.green}\nRich Presence utilized!{colors.reset}")
    league_of_legends_logo = "https://freepngimg.com/save/85643-blue-league-legends-icons-of-symbol-garena/1600x1600.png"
    start_time = int(time.time())

    # random_choices_in_client = [
    #    "In Lobby",
    #    "Navigating The Store",
    #    "Preparing for Game",
    #    "Locking in Positions",
    #    "Texting..",
    #    "Configuring settings",
    #    "In Lobby (3/5)" "In Lobby (4/5)" "In Lobby (Full)" "Looting Chests",
    #    "Bored..",
    # ]
    # random_choices_in_game = []

    while True:
        try:
            match player_state():
                case "InGame":
                    c_champion_name = get_champion_name()
                    c_champion_asset = champion_asset(champion=c_champion_name)
                    while player_state() == "InGame":
                        rpc.update(
                            large_image=c_champion_asset,
                            large_text=c_champion_name,
                            details=get_game_mode(),
                            state=f"In Game {get_kda()}",
                            small_image=league_of_legends_logo,
                            small_text="Smurfing in your peak.",
                            start=int(time.time())
                            - get_current_ingame_time(default_time=start_time),
                        )
                        time.sleep(10)
                case "InLobby":
                    rpc.update(
                        large_image=league_of_legends_logo,
                        large_text="In Lobby",
                        details="What the hell am i suppose to write here?",
                        state="In Lobby",
                        start=start_time,
                    )
                    time.sleep(10)

                case _:
                    print(
                        f"{colors.red}LeagueOfLegends.exe was terminated. rpc shuting down..."
                    )
                    sys.exit()
        except pypresence.exceptions.PipeClosed:
            print(
                f"{colors.red} Discord seems to be closed. Reconnect, and restart this script. {colors.reset}"
            )
            sys.exit()


if __name__ == "__main__":
    main(colors=Colors())
