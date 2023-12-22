import glob
import os
import sys
import time

import psutil
import pypresence

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
    print(f"{Colors.yellow}Checking if LeagueClient.exe is running...")
    time.sleep(1)

    # psutil is unable to find LeagueClientUx.exe (in my case?), instead it finds a process with name CrBrowserMain
    league_processes = ["LeagueClient.exe", "LeagueClientUx.exe", "CrBrowserMain"]

    if not processes_exists(process_names=league_processes):
        print(f"{Colors.red}League client is not running!{Colors.reset}")
        sys.exit()

    print(f"{Colors.green}League client is running!{Colors.dgray}(2/2){Colors.reset}")


def get_process_command(process_name: str) -> str:
    """
    Get process command of a process.
    """
    for proc in psutil.process_iter():
        try:
            if process_name.lower() in proc.name().lower():
                return proc.cmdline()
        except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
            pass
    return False


def get_lcu_connect_details(proc_name: str) -> [str, str]:
    """
    Get LCU connect details from command.
    """

    cmd = get_process_command(proc_name)

    for arg in cmd:
        if "--app-port" in arg:
            port = arg.split("=")[1]
        if "--remoting-auth-token" in arg:
            token = arg.split("=")[1]
    return port, token


def check_discord_process(
    process_names: list[str],
    client_id: str,
) -> pypresence.Presence:
    """
    Checks if discord process is running.
    Connects to Discord Rich Presence if it is found.
    """
    print(f"{Colors.yellow}Checking if Discord is running...{Colors.reset}")

    look_for_processes = f"({Colors.green}{', '.join(process_names)}{Colors.blue})"

    time.sleep(1)
    if not processes_exists(process_names=process_names):
        print(
            f"""{Colors.red}Discord not running!
            {Colors.blue}Could not find any process with the names {look_for_processes} running on your system.
            Is your Discord process named something else? Try --add-process <name>{Colors.reset}"""
        )
        sys.exit()

    print(f"{Colors.green}Discord is running! {Colors.dgray}(1/2){Colors.reset}")
    try:
        rpc = pypresence.Presence(client_id)
        rpc.connect()
    except ConnectionRefusedError:
        print(
            f"""
    {Colors.red}PyPresence encountered some problems, and could not connect to your Discord's RPC
    {Colors.blue}
    1. One or more of the processes this script was looking for was found {look_for_processes}
        But Pypresence still was unable to detect a running discord-ipc
    2. You may not have a discord ipc running. Try {Colors.reset}``{Colors.green}ls $XDG_RUNTIME_DIR | grep discord-ipc-{Colors.reset}``{Colors.blue} There should only be one result {Colors.reset}``{Colors.green}discord-ipc-0={Colors.reset}``
    {Colors.blue}3. Try restarting Discord. (Make sure the process is stopped before doing that.){Colors.reset}
            """
        )
        # If process names were not found, but ipc exists. Try removing them & restarting
        if len((val := check_discord_ipc())) > 1:
            print(
                f"""
{Colors.red}Detected multiple ipc's running.{Colors.reset}
    You seem to have more than 1 ipc running (which is unusual).
    If you know that discord is running, but pypresence keep failing to connect.
    It might be cause you have multiple ipc's running. try removing the following ipc's and {Colors.green}restart discord.{Colors.reset}
    {Colors.yellow}ipc's: {' , '.join(val)}{Colors.reset}
    run: ``{Colors.green}rm  {' '.join(val)}{Colors.reset}``
    Or you just don't have discord up and running..
            """
            )
        print(
            f"{Colors.red}Raising Exception found by PyPresence, and exiting..{Colors.reset}"
        )
        sys.exit()
    except pypresence.exceptions.InvalidID:
        print(
            f"{Colors.red}Invalid Client ID. Make sure your Discord Application ID is correct."
        )
        sys.exit()
    return rpc


def check_discord_ipc() -> list[str]:
    """
    Checks if there are any discord-ipc's running.
    """
    # Path to check for Discord IPC sockets

    xdg_runtime_dir = os.environ.get("XDG_RUNTIME_DIR")

    if not xdg_runtime_dir:
        # If for some reason this environmental variable is not set.. just ignore this function.
        return []

    ipc_pattern = "discord-ipc-*"
    list_of_ipcs: list[str] = []

    for ipc_socket in glob.glob(os.path.join(xdg_runtime_dir, ipc_pattern)):
        if os.path.exists(ipc_socket):
            list_of_ipcs.append(ipc_socket)
    return list_of_ipcs


def player_state() -> str | None:
    """
    Returns the player state
    """
    current_state: str | None = None

    if processes_exists(
        process_names=["LeagueClient.exe", "LeagueClientUx.exe", "CrBrowserMain"]
    ):
        if process_exists("League of Legends.exe"):
            current_state = "InGame"
        else:
            current_state = "InLobby"
    return current_state
