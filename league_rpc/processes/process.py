import sys
import threading
import time
from argparse import Namespace

import psutil
import pypresence  # type:ignore

from league_rpc.disable_native_rpc.disable import remove_plugin, find_game_path
from league_rpc.logger.richlogger import RichLogger
from league_rpc.utils.color import Color
from league_rpc.utils.launch_league import launch_league_client


def disable_native_presence() -> None:
    """Disable the native presence of League of Legends."""

    start_time = time.time()
    duration = 60  # 1 minute

    while True:
        if game_path := find_game_path():
            remove_plugin(file_path=game_path)

        if time.time() - start_time >= duration:
            # After 1 minute, stop the process.
            break


def processes_exists(process_names: list[str]) -> bool:
    """
    Given an array of process names.
    Give a boolean return value if any of the names was a running process in the machine.
    """
    return any(process_exists(process_name) for process_name in process_names)


def process_exists(process_name: str) -> bool:
    """
    Checks if the given process name is running or not.
    """
    for proc in psutil.process_iter():
        try:
            if process_name.lower() in proc.name().lower():
                return True
        except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
            pass
    return False


def check_league_client_process(cli_args: Namespace, logger: RichLogger) -> None:
    """
    Checks league client processes.
    """
    league_processes: list[str] = ["LeagueClient.exe", "LeagueClientUx.exe"]

    logger.info("Checking if LeagueClient.exe is running...", color="yellow")
    time.sleep(1)
    logger.update_progress_bar(advance=40)

    if cli_args.launch_league:
        # launch league if it's not already running.
        if not processes_exists(league_processes):
            disable_native_thread = threading.Thread(
                target=disable_native_presence,
                daemon=True,
            )
            disable_native_thread.start()
            launch_league_client(cli_args)

            time.sleep(0.5)
            logger.info("League Client has been launched!")
            logger.info("Disabling the Native league presence", color="yellow")

    if not processes_exists(process_names=league_processes):
        # If league process is still not running, even after launching the client.
        # Then something must have gone wrong.
        # Do not exit app, but rather wait for user to open the correct game..

        if cli_args.wait_for_league == -1:
            logger.warning(
                "Will wait indefinitely for League to start... Remember, forever is a long time.. use CTRL + C if you would like to quit.",
                color="yellow",
                highlight=[
                    {
                        "CTRL + C": "green",
                        "indefinitely": "green",
                    }
                ],
            )

    wait_time = 0
    while not processes_exists(process_names=league_processes):
        if cli_args.wait_for_league == -1:
            continue
        elif wait_time >= cli_args.wait_for_league:
            logger.error(
                f"League Client is not running! Exiting after waiting {cli_args.wait_for_league} seconds."
            )

            if not cli_args.wait_for_league:
                logger.info(
                    "Want to add waiting time for League? Use --wait-for-league <seconds>. (-1 = infinite, or until CTRL + C)",
                    color="green",
                )

            sys.exit()
        else:
            logger.info(
                f"Will wait for League to start. Time left: {cli_args.wait_for_league - wait_time} seconds...",
                color="yellow",
            )

            time.sleep(5)
            wait_time += 5
            continue
        break

    logger.info("League client is running!", color="green")
    logger.update_progress_bar(advance=50)


def check_discord_process(
    process_names: list[str], client_id: str, wait_for_discord: int, logger: RichLogger
) -> pypresence.Presence:
    """
    Checks if discord process is running.
    Connects to Discord Rich Presence if it is found.
    """
    logger.info("Checking if Discord is running...", color="yellow")

    look_for_processes = f"({Color.green}{', '.join(process_names)}{Color.blue})"

    time.sleep(1)
    logger.update_progress_bar(advance=20)

    if not processes_exists(process_names=process_names):
        if wait_for_discord == -1:
            logger.warning(
                "Will wait indefinitely for Discord to start... Remember, forever is a long time.. use CTRL + C if you would like to quit.",
                color="yellow",
                highlight=[
                    {
                        "CTRL + C": "green",
                        "indefinitely": "green",
                    }
                ],
            )

    wait_time = 0
    while True:
        if not processes_exists(process_names=process_names):
            if wait_for_discord == -1:
                time.sleep(3)
                continue
            elif wait_time >= wait_for_discord:
                logger.error(
                    f"Discord not running! Could not find any process with the names {look_for_processes} running on your system."
                )
                logger.info(
                    "Is your Discord process named something else? Try --add-process <name>"
                )

                if not wait_for_discord:
                    logger.info(
                        "Want to add waiting time for discord? Use --wait-for-discord <seconds>. (-1 = infinite, or until CTRL + C)",
                        color="green",
                    )

                sys.exit()
            else:
                logger.info(
                    f"Will wait for Discord to start. Time left: {wait_for_discord - wait_time} seconds...",
                    color="yellow",
                )

                time.sleep(5)
                wait_time += 5
                continue
        break

    logger.info("Discord is running!", color="green")
    logger.update_progress_bar(advance=50)

    for _ in range(5):
        time.sleep(3)
        try:
            rpc = pypresence.Presence(client_id)
            rpc.connect()
            logger.info("Connected to Discord RPC!", color="green")
            logger.update_progress_bar(advance=30)
            break
        except pypresence.exceptions.InvalidID:
            logger.error(
                "Invalid Client ID. Make sure your Discord Application ID is correct."
            )

            sys.exit()
        except pypresence.exceptions.DiscordError:
            # Sometimes when starting discord, an error can occur saying that you logged out.
            # Weird but can be ignored since it usually works a second or so after.
            time.sleep(1)
            continue
        except pypresence.exceptions.DiscordNotFound:
            # Sometimes when starting discord, an error can occur saying that you logged out.
            # Weird but can be ignored since it usually works a second or so after.

            logger.error(
                "Discord not found. Make sure Discord is installed and running on your system."
            )

            sys.exit()
        except pypresence.exceptions.PipeClosed:
            # The pipe was closed. Catch this exception and re-connect your instance
            time.sleep(1)
            continue
        except ConnectionRefusedError:
            logger.error(
                "Connection refused. Make sure Discord is running and accepting connections."
            )
            logger.info(
                "Try to restart Discord (Close the application from your Task Manager)."
            )
            sys.exit()
    else:
        logger.error("Discord process was found but RPC could not be connected.")

        sys.exit()
    return rpc


def player_state() -> str | None:
    """
    Returns the player state
    """
    current_state: str | None = None

    if processes_exists(process_names=["LeagueClient.exe", "LeagueClientUx.exe"]):
        if process_exists(process_name="League of Legends.exe"):
            current_state = "InGame"
        else:
            current_state = "InLobby"
    return current_state
