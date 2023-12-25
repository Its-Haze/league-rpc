import argparse
import sys
import time
from multiprocessing import Process
import nest_asyncio

import pypresence

from league_rpc_linux.champion import gather_ingame_information, get_skin_asset
from league_rpc_linux.colors import Colors
from league_rpc_linux.gametime import get_current_ingame_time
from league_rpc_linux.kda import get_creepscore, get_gold, get_kda, get_level
from league_rpc_linux.processes.process import (
    check_discord_process,
    check_league_client_process,
    player_state,
)

from league_rpc_linux.reconnect import discord_reconnect_attempt
from league_rpc_linux.processes import LCU_Thread

# Discord Application: League of Linux

from league_rpc_linux.const import (
    DEFAULT_CLIENT_ID,
    DISCORD_PROCESS_NAMES,
    LEAGUE_OF_LEGENDS_LOGO,
    SMALL_TEXT,
    CURRENT_PATCH,
)


def main(cli_args: argparse.Namespace):
    """
    This is the program that gets executed.
    """
    ############################################################
    ## Check Discord, RiotClient & LeagueClient processes     ##
    check_league_client_process(wait_for_league=cli_args.wait_for_league)

    rpc = check_discord_process(
        process_names=DISCORD_PROCESS_NAMES + cli_args.add_process,
        client_id=cli_args.client_id,
        wait_for_discord=cli_args.wait_for_discord,
    )

    check_league_client_process()

    p = Process(target=LCU_Thread.start_connector, args=(rpc,))
    p.start()

    print(f"\n{Colors.green}Successfully connected to Discord RPC!{Colors.reset}")
    ############################################################
    start_time = int(time.time())
    while True:
        try:
            match player_state():
                case "InGame":
                    print(
                        f"\n{Colors.dblue}Detected game! Will soon gather data and update discord RPC{Colors.reset}"
                    )
                    (
                        champ_name,
                        skin_name,
                        skin_id,
                        gamemode,
                        _,
                        _,
                    ) = gather_ingame_information()
                    if gamemode == "TFT":
                        # TFT RPC
                        while player_state() == "InGame":
                            rpc.update(  # type:ignore
                                large_image="https://wallpapercave.com/wp/wp7413493.jpg",
                                large_text="Playing TFT",
                                details="Teamfight Tactics",
                                state=f"In Game · lvl: {get_level()}",
                                small_image=LEAGUE_OF_LEGENDS_LOGO,
                                small_text=SMALL_TEXT,
                                start=int(time.time())
                                - get_current_ingame_time(default_time=start_time),
                            )
                            time.sleep(10)
                    elif gamemode == "Arena":
                        # ARENA RPC
                        skin_asset = get_skin_asset(
                            champion_name=champ_name,
                            skin_id=skin_id,
                        )
                        print(
                            f"{Colors.green}Successfully gathered all data.{Colors.yellow}\nUpdating Discord Presence now!{Colors.reset}"
                        )
                        while player_state() == "InGame":
                            rpc.update(  # type:ignore
                                large_image=skin_asset,
                                large_text=skin_name if skin_name else champ_name,
                                details=gamemode,
                                state=f"In Game {f'· {get_kda()} · lvl: {get_level()} · gold: {get_gold()}' if not cli_args.no_stats else ''}",
                                small_image=LEAGUE_OF_LEGENDS_LOGO,
                                small_text=SMALL_TEXT,
                                start=int(time.time())
                                - get_current_ingame_time(default_time=start_time),
                            )
                            time.sleep(10)
                    else:
                        # LEAGUE RPC
                        skin_asset = get_skin_asset(
                            champion_name=champ_name,
                            skin_id=skin_id,
                        )
                        print(
                            f"{Colors.green}Successfully gathered all data.{Colors.yellow}\nUpdating Discord Presence now!{Colors.reset}"
                        )
                        while player_state() == "InGame":
                            if not champ_name or not gamemode:
                                print(
                                    f"{Colors.red}Failed to load in data.. {Colors.lgrey}will try again shortly.\n{Colors.dcyan}(Reason: Someone has potato PC, meaning that RITOs API isn't fully initialized yet but the script sees that game has started.){Colors.reset}"
                                )
                                break
                            rpc.update(  # type:ignore
                                large_image=skin_asset,
                                large_text=skin_name if skin_name else champ_name,
                                details=gamemode,
                                state=f"In Game {f'· {get_kda()} · {get_creepscore()}' if not cli_args.no_stats else ''}",
                                small_image=LEAGUE_OF_LEGENDS_LOGO,
                                small_text=SMALL_TEXT,
                                start=int(time.time())
                                - get_current_ingame_time(default_time=start_time),
                            )
                            time.sleep(10)

                case "InLobby":
                    # handled by LCU_Thread

                    time.sleep(10)

                case _:
                    print(
                        f"{Colors.red}LeagueOfLegends.exe was terminated. rpc shuting down..{Colors.reset}."
                    )
                    p.terminate()
                    rpc.close()
                    sys.exit()
        except pypresence.exceptions.PipeClosed:
            # If the program crashes because pypresence failed to connect to a pipe. (Typically if Discord is closed.)
            # The script will automatically try to reconnect..
            # if it fails it will keep going until you either reconnect or after a long enough period of time has passed
            print(
                f"{Colors.red}Discord seems to be closed, will attempt to reconnect!{Colors.reset}"
            )
            discord_reconnect_attempt(rpc, amount_of_tries=12, amount_of_waiting=5)


if __name__ == "__main__":
    # Patch for asyncio - read more here: https://pypi.org/project/nest-asyncio/
    nest_asyncio.apply()

    parser = argparse.ArgumentParser(description="Script with Discord RPC.")
    parser.add_argument(
        "--client-id",
        type=str,
        default=DEFAULT_CLIENT_ID,
        help=f"Client ID for Discord RPC. Default is {DEFAULT_CLIENT_ID}. which will show 'League of Linux' on discord",
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
    parser.add_argument(
        "--wait-for-league",
        type=int,
        default=0,
        help="Time in seconds to wait for the League client to start. -1 for infinite waiting, Good when used as a starting script for league.",
    )
    parser.add_argument(
        "--wait-for-discord",
        type=int,
        default=0,
        help="Time in seconds to wait for the Discord client to start. -1 for infinite waiting, Good when you want to start this script before you've had time to start Discord.",
    )

    args = parser.parse_args()

    # Prints the League RPC logo
    print(Colors().logo)

    if args.no_stats:
        print(
            f"{Colors.green}Argument {Colors.blue}--no-stats{Colors.green} detected.. Will {Colors.red}not {Colors.green}show InGame stats{Colors.reset}"
        )
    if args.add_process:
        print(
            f"{Colors.green}Argument {Colors.blue}--add-process{Colors.green} detected.. Will add {Colors.blue}{args.add_process}{Colors.green} to the list of Discord processes to look for.{Colors.reset}"
        )
    if args.client_id != DEFAULT_CLIENT_ID:
        print(
            f"{Colors.green}Argument {Colors.blue}--client-id{Colors.green} detected.. Will try to connect by using {Colors.blue}({args.client_id}){Colors.reset}"
        )
    if args.wait_for_league:
        print(
            f"{Colors.green}Argument {Colors.blue}--wait-for-league{Colors.green} detected.. {Colors.blue}will wait for League to start before continuing{Colors.reset}"
        )
    if args.wait_for_discord:
        print(
            f"{Colors.green}Argument {Colors.blue}--wait-for-discord{Colors.green} detected.. {Colors.blue}will wait for Discord to start before continuing{Colors.reset}"
        )

    main(args)
