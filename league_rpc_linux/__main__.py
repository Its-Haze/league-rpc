import argparse
import sys
import time

import pypresence

from league_rpc_linux.champion import gather_ingame_information, get_skin_asset
from league_rpc_linux.colors import Colors
from league_rpc_linux.gametime import get_current_ingame_time
from league_rpc_linux.kda import get_kda
from league_rpc_linux.processes.process import (
    check_discord_process,
    check_league_client_process,
    check_riot_games_service_process,
    player_state,
)

# Discord Application: League of Linux
DEFAULT_CLIENT_ID = "1185274747836174377"
DISCORD_PROCESS_NAMES = ["Discord", "DiscordPTB", "DiscordCanary", "electron"]


parser = argparse.ArgumentParser(description="Script with Discord RPC.")
parser.add_argument(
    "--client-id",
    type=str,
    default=DEFAULT_CLIENT_ID,
    help="Client ID for Discord RPC. Default is 1185274747836174377. which will show 'League of Linux' on discord",
)
parser.add_argument(
    "--no-stats",
    action="store_true",
    help="use '--no-stats' to Opt out of showing in-game stats (KDA, minions) in Discord RPC",
)
parser.add_argument(
    "--add-process",
    nargs="+",
    default=[],
    help="Add custom Discord process names to the search list.",
)


args = parser.parse_args()


def main():
    """
    main program.
    """
    print(Colors().logo)

    if args.no_stats:
        print(
            f"{Colors.green}Argument --no-stats detected.. Will {Colors.red}not {Colors.green}show InGame stats{Colors.reset}"
        )
    if args.add_process:
        print(
            f"{Colors.green}Argument --add-process detected.. Will add {Colors.blue}{args.add_process}{Colors.green} to the list of Discord processes to look for.{Colors.reset}"
        )

    rpc = check_discord_process(
        process_names=DISCORD_PROCESS_NAMES + args.add_process, client_id=args.client_id
    )

    check_riot_games_service_process()
    check_league_client_process()

    print(f"{Colors.green}\nRich Presence utilized!{Colors.reset}")
    league_of_legends_logo = "https://freepngimg.com/save/85643-blue-league-legends-icons-of-symbol-garena/1600x1600.png"
    start_time = int(time.time())

    while True:
        try:
            match player_state():
                case "InGame":
                    print(
                        f"{Colors.dblue}Detected game! Will soon gather data and update discord RPC{Colors.reset}"
                    )

                    (
                        champ_name,
                        skin_id,
                        gamemode,
                        _,
                    ) = gather_ingame_information()
                    if gamemode != "TFT":
                        # LEAGUE RPC
                        skin_asset = get_skin_asset(
                            champion_name=champ_name, skin_id=skin_id
                        )
                        while player_state() == "InGame":
                            rpc.update(  # type:ignore
                                large_image=skin_asset,
                                large_text=champ_name,
                                details=gamemode,
                                state=f"In Game {get_kda() if not args.no_stats else ''}",
                                small_image=league_of_legends_logo,
                                small_text="github.com/Its-Haze/league-rpc-linux",
                                start=int(time.time())
                                - get_current_ingame_time(default_time=start_time),
                            )
                            if champ_name == "???" or gamemode == "???":
                                print(
                                    f"{Colors.red}Failed to load in data.. {Colors.lgrey}will try again shortly.\n{Colors.dcyan}(Reason: Someone has potato PC, meaning that RITOs API isn't fully initialized yet but the script sees that game has started.){Colors.reset}"
                                )
                                break
                            time.sleep(10)
                    else:
                        # TFT RPC
                        while player_state() == "InGame":
                            (
                                _,
                                _,
                                gamemode,
                                level,
                            ) = gather_ingame_information()
                            if gamemode and level:
                                rpc.update(  # type:ignore
                                    large_image="https://wallpapercave.com/wp/wp7413493.jpg",
                                    large_text="Playing TFT",
                                    details="Teamfight Tactics",
                                    state=f"In Game · lvl: {level}",
                                    small_image=league_of_legends_logo,
                                    small_text="github.com/Its-Haze/league-rpc-linux",
                                    start=int(time.time())
                                    - get_current_ingame_time(default_time=start_time),
                                )
                            time.sleep(10)

                case "InLobby":
                    rpc.update(  # type:ignore
                        large_image=league_of_legends_logo,
                        large_text="github.com/Its-Haze/league-rpc-linux",
                        state="In Client",
                        start=start_time,
                    )
                    time.sleep(10)

                case _:
                    print(
                        f"{Colors.red}LeagueOfLegends.exe was terminated. rpc shuting down..{Colors.reset}."
                    )
                    rpc.close()
                    sys.exit()
        except pypresence.exceptions.PipeClosed:
            # If the program crashes because pypresence failed to connect to a pipe. (Typically if Discord is closed.)
            # The script will automatically try to reconnect..
            # if it fails it will keep going until you either reconnect or after a long enough period of time has passed
            print(
                f"{Colors.red}Discord seems to be closed, will attempt to reconnect!{Colors.reset}"
            )
            amount_of_tries = 10
            amount_of_waiting = 5

            for i in range(amount_of_tries):
                try:
                    time.sleep(amount_of_waiting)
                    print(f"{Colors.yellow}{i}. Attempting to reconnect..{Colors.reset}")
                    rpc = pypresence.Presence(args.client_id)
                    rpc.connect()
                    print(
                        f"{Colors.green}Successfully reconnected.. Proceeding as normal.{Colors.reset}"
                    )
                    break

                except Exception:  # pylint:disable=broad-exception-caught
                    print(
                        f"{Colors.red}Exception caught but continuing..{Colors.reset}"
                    )
            else:
                print(
                    f"{Colors.red}Was unable to reconnect to discord. after trying for {amount_of_tries * amount_of_waiting} seconds.{Colors.reset}"
                )
                sys.exit()


if __name__ == "__main__":
    main()
