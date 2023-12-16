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


def check_league_client_process(colors: Colors):
    """
    Checks league client processes.
    """
    print(colors.yellow + "Checking if LeagueClient.exe is running...")
    time.sleep(1)

    league_processes = ["LeagueClient.exe", "LeagueClientUx.exe"]

    if processes_exists(process_names=league_processes):
        print(
            colors.green
            + "League client is running!"
            + colors.dgray
            + "(3/3)"
            + colors.reset
        )
    else:
        print(colors.red + "League client is not running!" + colors.reset)
        time.sleep(1)
        exit()


def check_riot_games_service_process(colors: Colors):
    """
    Checks that the Riot Games launcher is running.
    """
    print(colors.yellow + "Checking if Riot Games Launcher is running...")
    time.sleep(2)
    if process_exists("RiotClientServi"):
        print(
            colors.green
            + "Riot Games Service is running!"
            + colors.dgray
            + "(2/3)"
            + colors.reset
        )
    else:
        print(colors.red + "Riot Games Service is not running!" + colors.reset)
        time.sleep(1)
        exit()


def check_discord_process(colors: Colors, process_names: list[str]) -> Presence:
    """
    Checks if discord process is running.
    Connects to Discord Rich Presence if it is found.
    """

    print(colors.yellow + "Checking if Discord is running...")
    time.sleep(1)
    if processes_exists(process_names=process_names):
        print(
            colors.green + "Discord is running!" + colors.dgray + "(1/3)" + colors.reset
        )
        rpc = Presence("1185274747836174377")
        # League of linux: 1185274747836174377
        # League of Legends: 401518684763586560

        rpc.connect()
    else:
        print(colors.red + "Discord not running!" + colors.reset)
        exit()
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
