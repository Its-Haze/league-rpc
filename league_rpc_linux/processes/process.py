import sys
import time

import psutil
from pypresence import Presence

from league_rpc_linux.colors import Colors


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


def check_league_client_process():
    """
    Checks league client processes.
    """
    print(Colors.yellow + "Checking if LeagueClient.exe is running...")
    time.sleep(1)

    league_processes = ["LeagueClient.exe", "LeagueClientUx.exe"]

    if processes_exists(process_names=league_processes):
        print(
            Colors.green
            + "League client is running!"
            + Colors.dgray
            + "(3/3)"
            + Colors.reset
        )
    else:
        print(Colors.red + "League client is not running!" + Colors.reset)
        time.sleep(1)
        sys.exit()


def check_riot_games_service_process() -> None:
    """
    Checks that the Riot Games launcher is running.
    """
    print(Colors.yellow + "Checking if Riot Games Launcher is running...")
    time.sleep(2)
    if process_exists("RiotClientServi"):
        print(
            Colors.green
            + "Riot Games Service is running!"
            + Colors.dgray
            + "(2/3)"
            + Colors.reset
        )
    else:
        print(Colors.red + "Riot Games Service is not running!" + Colors.reset)
        time.sleep(1)
        sys.exit()


def check_discord_process(process_names: list[str]) -> Presence:
    """
    Checks if discord process is running.
    Connects to Discord Rich Presence if it is found.
    """

    print(Colors.yellow + "Checking if Discord is running...")
    time.sleep(1)
    if processes_exists(process_names=process_names):
        print(
            Colors.green + "Discord is running!" + Colors.dgray + "(1/3)" + Colors.reset
        )
        rpc = Presence("1185274747836174377")
        # League of linux: 1185274747836174377
        # League of Legends: 401518684763586560

        rpc.connect()
    else:
        print(Colors.red + "Discord not running!" + Colors.reset)
        sys.exit()
    return rpc


def player_state() -> str | None:
    """
    Returns the player state
    """
    current_state: str | None = None

    if process_exists("RiotClientServi"):
        if process_exists("LeagueClient.exe") or process_exists("LeagueClientUx.exe"):
            if process_exists("League of Legends.exe"):
                current_state = "InGame"
            else:
                current_state = "InLobby"
    return current_state
