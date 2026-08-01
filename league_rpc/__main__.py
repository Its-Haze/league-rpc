import argparse
import random
import sys
import threading
import time

import nest_asyncio  # type:ignore
import pypresence  # type:ignore

from league_rpc.lcu_api.lcu_connector import module_data, start_connector
from league_rpc.logger.richlogger import RichLogger
from league_rpc.processes.process import (
    check_discord_process,
    check_league_client_process,
)
from league_rpc.utils.color import Color
from league_rpc.utils.const import (
    ANIMATED_SKIN_URL,
    ANIMATED_SKINS,
    DEFAULT_CLIENT_ID,
    DISCORD_PROCESS_NAMES,
    LEAGUE_CLASSIC_ICON,
    PLACEHOLDER_ICON_ROTATE_INTERVAL_SECONDS,
    SMALL_TEXT,
)
from league_rpc.utils.launch_league import find_default_path


def _random_animated_skin() -> str:
    return ANIMATED_SKIN_URL.format(filename=random.choice(ANIMATED_SKINS))


def main(cli_args: argparse.Namespace) -> None:
    """
    This is the program that gets executed.
    """

    logger = RichLogger(show_debugs=cli_args.debug)

    ############################################################
    ## Check Discord, RiotClient & LeagueClient processes     ##

    # Connect to Discord before checking/launching League, so we claim the Rich
    # Presence slot before League's own client gets a chance to connect.
    logger.start_progress_bar(name="Checking Discord")
    rpc = check_discord_process(
        process_names=DISCORD_PROCESS_NAMES + cli_args.add_process,
        client_id=cli_args.client_id,
        wait_for_discord=cli_args.wait_for_discord,
        logger=logger,
    )
    logger.stop_progress_bar()

    placeholder_start = int(time.time())

    def send_placeholder() -> None:
        with module_data.rpc_lock:
            try:
                rpc.update(
                    large_image=_random_animated_skin(),
                    large_text="LeagueRPC",
                    small_image=LEAGUE_CLASSIC_ICON,
                    small_text=SMALL_TEXT,
                    details="Launching League...",
                    state="LeagueRPC",
                    start=placeholder_start,
                )
            except pypresence.exceptions.PyPresenceException:
                pass

    def rotate_placeholder_icon(stop_event: threading.Event) -> None:
        while not stop_event.wait(PLACEHOLDER_ICON_ROTATE_INTERVAL_SECONDS):
            send_placeholder()

    send_placeholder()

    stop_rotation = threading.Event()
    threading.Thread(
        target=rotate_placeholder_icon, args=(stop_rotation,), daemon=True
    ).start()

    logger.start_progress_bar(name="Checking League")
    check_league_client_process(cli_args, logger)
    logger.stop_progress_bar()

    stop_rotation.set()

    # Start LCU_Thread
    # This process will connect to the LCU API and updates the rpc based on data subscribed from the LCU API.
    # In this case passing the rpc object to the process is easier than trying to return updated data from the process.
    # Every In-Client update will be handled by the LCU_Thread process and will update the rpc accordingly.
    lcu_process = threading.Thread(
        target=start_connector,
        args=(
            rpc,
            cli_args,
            logger,
        ),
        daemon=True,
    )
    lcu_process.start()

    try:
        while lcu_process.is_alive():
            time.sleep(1)
    except KeyboardInterrupt as e:
        logger.info(f"{e.__class__.__name__} detected. Shutting down the program..")
    finally:
        # Always close the RPC connection when exiting (whether by KeyboardInterrupt or League closing).
        # Discord may have already closed the IPC pipe on its end (e.g. Discord was
        # closed, or the pipe timed out) by the time we get here, which pypresence
        # surfaces as PyPresenceException (PipeClosed, ResponseTimeout, ...), or as a
        # RuntimeError if the heartbeat/reclaim threads were mid-call on the same
        # connection. This is best-effort cleanup on the way out, so nothing here
        # should crash the app - hence the broad except.
        with module_data.rpc_lock:
            try:
                rpc.clear()
            except Exception:
                pass
            try:
                rpc.close()
            except Exception:
                pass
        logger.info("Discord RPC connection closed.")

    ############################################################


if __name__ == "__main__":
    # Patch for asyncio - read more here: https://pypi.org/project/nest-asyncio/
    nest_asyncio.apply()  # type: ignore

    default_league_path = find_default_path()

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
        "--hide-emojis",
        action="store_true",
        help="use '--hide-emojis' to hide the green/red circle emoji, depending on your Online status in league.",
    )
    parser.add_argument(
        "--debug",
        action="store_true",
        help="use '--debug' to see additional logging data. Otherwise logs from INFO -> CRITICAL will show. Debug logs are hidden by default",
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
    parser.add_argument(
        "--launch-league",
        type=str,
        default=default_league_path,
        help=f"Path to the League of Legends client executable. Default path is: {default_league_path}",
    )
    parser.add_argument(
        "--hide-in-client",
        action="store_true",
        help="Temporarily hides the League of Legends Rich Presence while you're idle in the client only. The presence will automatically reactivate when you enter a lobby, queue, champ select, or game.",
    )

    args: argparse.Namespace = parser.parse_args()

    # Prints the League RPC logo
    print(Color().logo)

    # Discord server information for support and other projects
    print(f"{Color.blue}{'─' * 60}{Color.reset}")
    print(
        f"{Color.green}• Need help? Join the Discord: {Color.cyan}https://discord.haze.sh{Color.reset}"
    )
    print(
        f"{Color.green}• Like LeagueRPC? You'll love {Color.orange}DJ Braum{Color.green} - Music bot for you & your friends: {Color.cyan}https://braum.haze.sh{Color.reset}"
    )
    print(
        f"{Color.green}• Portfolio & more projects: {Color.cyan}https://haze.sh{Color.reset} {Color.green}Be careful of the 🦆. It might get angry!{Color.reset}"
    )
    print(f"{Color.blue}{'─' * 60}{Color.reset}\n")

    if args.hide_in_client:
        print(
            f"{Color.green}Argument {Color.blue}--hide-in-client{Color.green} detected.. Will hide the in-client Rich presence.{Color.reset}"
        )
    if args.no_stats:
        print(
            f"{Color.green}Argument {Color.blue}--no-stats{Color.green} detected.. Will {Color.red}not {Color.green}show InGame stats{Color.reset}"
        )
    if args.no_rank:
        print(
            f"{Color.green}Argument {Color.blue}--no-rank{Color.green} detected.. Will hide your league rank.{Color.reset}"
        )
    if args.hide_emojis:
        print(
            f"{Color.green}Argument {Color.blue}--hide-emojis{Color.green} detected.. Will hide emojis. such as league status indicators on Discord.{Color.reset}"
        )
    if args.debug:
        print(
            f"{Color.green}Argument {Color.blue}--debug{Color.green} detected.. Will show debug logs.{Color.reset}"
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

    if args.launch_league:
        if args.launch_league == default_league_path:
            print(
                f"{Color.green}Attempting to launch the League client at the default location{Color.reset} {Color.blue}{default_league_path}{Color.reset}\n"
                f"{Color.green}If league is already running, it will not launch a new instance.{Color.reset}\n"
                f"{Color.orange}If the League client does not launch, please specify the path manually using: --launch-league <path>{Color.reset}\n"
            )
        else:
            print(
                f"{Color.green}Detected the {Color.blue}--launch-league{Color.green} argument with a custom path. Attempting to launch the League client, from: {Color.blue}{args.launch_league}{Color.reset}\n"
                f"{Color.orange}If league is already running, it will not launch a new instance.{Color.reset}\n"
            )

    main(cli_args=args)
