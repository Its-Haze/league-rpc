import argparse
import sys
import threading
import time

import nest_asyncio
import pypresence

from league_rpc.champion import gather_ingame_information, get_skin_asset
from league_rpc.utils.color import Color
from league_rpc.utils.const import (
    ALL_GAME_DATA_URL,
    CHAMPION_NAME_CONVERT_MAP,
    DEFAULT_CLIENT_ID,
    DISCORD_PROCESS_NAMES,
    LEAGUE_OF_LEGENDS_LOGO,
    SMALL_TEXT,
)
from league_rpc.gametime import get_current_ingame_time
from league_rpc.kda import get_creepscore, get_gold, get_kda, get_level
from league_rpc.lcu_api.lcu_connector import start_connector
from league_rpc.utils.polling import wait_until_exists
from league_rpc.processes.process import (
    check_discord_process,
    check_league_client_process,
    player_state,
)
from league_rpc.reconnect import discord_reconnect_attempt

# Discord Application: League of Linux


def main(cli_args: argparse.Namespace) -> None:
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

    # Start LCU_Thread
    # This process will connect to the LCU API and updates the rpc based on data subscribed from the LCU API.
    # In this case passing the rpc object to the process is easier than trying to return updated data from the process.
    # Every In-Client update will be handled by the LCU_Thread process and will update the rpc accordingly.
    lcu_process = threading.Thread(
        target=start_connector,
        args=(
            rpc,
            cli_args,
        ),
        daemon=True,
    )
    lcu_process.start()

    print(f"\n{Color.green}Successfully connected to Discord RPC!{Color.reset}")
    ############################################################

    start_time = int(time.time())
    while True:
        try:
            match player_state():
                case "InGame":
                    print(
                        f"\n{Color.dblue}Detected game! Will soon gather data and update discord RPC{Color.reset}"
                    )

                    # Poll the local league api until 200 response.
                    wait_until_exists(
                        url=ALL_GAME_DATA_URL,
                        custom_message="Failed to reach the local league api",
                        startup=True,
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
                        skin_asset: str = get_skin_asset(
                            champion_name=champ_name,
                            skin_id=skin_id,
                        )
                        print(
                            f"{Color.green}Successfully gathered all data.{Color.yellow}\nUpdating Discord Presence now!{Color.reset}"
                        )
                        while player_state() == "InGame":
                            rpc.update(  # type:ignore
                                large_image=skin_asset,
                                large_text=(
                                    skin_name
                                    if skin_name
                                    else CHAMPION_NAME_CONVERT_MAP.get(
                                        champ_name, champ_name
                                    )
                                ),
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
                            f"{Color.green}Successfully gathered all data.{Color.yellow}\nUpdating Discord Presence now!{Color.reset}"
                        )
                        while player_state() == "InGame":
                            if not champ_name or not gamemode:
                                break
                            rpc.update(  # type:ignore
                                large_image=skin_asset,
                                large_text=(
                                    skin_name
                                    if skin_name
                                    else CHAMPION_NAME_CONVERT_MAP.get(
                                        champ_name, champ_name
                                    )
                                ),
                                details=gamemode,
                                state=f"In Game {f'· {get_kda()} · {get_creepscore()}' if not cli_args.no_stats else ''}",
                                small_image=LEAGUE_OF_LEGENDS_LOGO,
                                small_text=SMALL_TEXT,
                                start=int(time.time())
                                - get_current_ingame_time(default_time=start_time),
                            )
                            time.sleep(10)

                case "InLobby":
                    # Handled by lcu_process thread
                    # It will subscribe to websockets and update discord on events.

                    time.sleep(10)

                case _:
                    print(
                        f"{Color.red}LeagueOfLegends.exe was terminated. rpc shuting down..{Color.reset}."
                    )
                    rpc.close()
                    sys.exit()
        except pypresence.exceptions.PipeClosed:
            # If the program crashes because pypresence failed to connect to a pipe. (Typically if Discord is closed.)
            # The script will automatically try to reconnect..
            # if it fails it will keep going until you either reconnect or after a long enough period of time has passed
            print(
                f"{Color.red}Discord seems to be closed, will attempt to reconnect!{Color.reset}"
            )
            discord_reconnect_attempt(rpc=rpc, amount_of_tries=12, amount_of_waiting=5)


if __name__ == "__main__":
    # Patch for asyncio - read more here: https://pypi.org/project/nest-asyncio/
    nest_asyncio.apply()  # type: ignore

    parser = argparse.ArgumentParser(description="Script with Discord RPC.")
    parser.add_argument(
        "--client-id",
        type=str,
        default=DEFAULT_CLIENT_ID,
        help=f"Client ID for Discord RPC. Default is {DEFAULT_CLIENT_ID}. which will show 'League of Legends' on Discord",
    )
    parser.add_argument(
        "--no-stats",
        action="store_true",
        help="use '--no-stats' to Opt out of showing in-game stats (KDA, minions) in Discord RPC",
    )
    parser.add_argument(
        "--show-emojis",
        "--emojis",
        action="store_true",
        help="use '--show-emojis' to show green/red circle emoji, depending on your Online status in league.",
    )
    parser.add_argument(
        "--no-rank",
        action="store_true",
        help="use '--no-rank' to hide your SoloQ/Flex/Tft/Arena Rank in Discord RPC",
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
        default=-1,
        help="Time in seconds to wait for the League client to start. -1 for infinite waiting, Good when used as a starting script for league.",
    )
    parser.add_argument(
        "--wait-for-discord",
        type=int,
        default=-1,
        help="Time in seconds to wait for the Discord client to start. -1 for infinite waiting, Good when you want to start this script before you've had time to start Discord.",
    )

    args: argparse.Namespace = parser.parse_args()

    # Prints the League RPC logo
    print(Color().logo)

    if args.no_stats:
        print(
            f"{Color.green}Argument {Color.blue}--no-stats{Color.green} detected.. Will {Color.red}not {Color.green}show InGame stats{Color.reset}"
        )
    if args.no_rank:
        print(
            f"{Color.green}Argument {Color.blue}--no-rank{Color.green} detected.. Will hide your league rank.{Color.reset}"
        )
    if args.show_emojis:
        print(
            f"{Color.green}Argument {Color.blue}--show-emojis, --emojis{Color.green} detected.. Will show emojis. such as league status indicators on Discord.{Color.reset}"
        )
    if args.add_process:
        print(
            f"{Color.green}Argument {Color.blue}--add-process{Color.green} detected.. Will add {Color.blue}{args.add_process}{Color.green} to the list of Discord processes to look for.{Color.reset}"
        )
    if args.client_id != DEFAULT_CLIENT_ID:
        print(
            f"{Color.green}Argument {Color.blue}--client-id{Color.green} detected.. Will try to connect by using {Color.blue}({args.client_id}){Color.reset}"
        )
    if args.wait_for_league and args.wait_for_league > 0:
        print(
            f"{Color.green}Argument {Color.blue}--wait-for-league{Color.green} detected.. {Color.blue}will wait for League to start before continuing{Color.reset}"
        )
    if args.wait_for_discord and args.wait_for_discord > 0:
        print(
            f"{Color.green}Argument {Color.blue}--wait-for-discord{Color.green} detected.. {Color.blue}will wait for Discord to start before continuing{Color.reset}"
        )

    main(cli_args=args)
