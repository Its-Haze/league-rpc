import os
import subprocess
from argparse import Namespace
from typing import LiteralString

from league_rpc.utils.const import (
    COMMON_DRIVES,
    DEFAULT_LEAGUE_CLIENT_EXE_PATH,
    DEFAULT_LEAGUE_CLIENT_EXECUTABLE,
)


def find_default_path() -> str | LiteralString:
    """Find the default path of the League Client executable."""

    # Check each drive for the executable
    for drive in COMMON_DRIVES:
        full_path = os.path.join(drive, DEFAULT_LEAGUE_CLIENT_EXE_PATH)
        if os.path.exists(full_path):
            return full_path
    # Fallback if the executable is not found
    return os.path.join("C:", DEFAULT_LEAGUE_CLIENT_EXE_PATH)


def launch_league_client(cli_args: Namespace) -> None:
    """Launch the League Client with the given path or the default path."""

    # If the user wants to launch the league client.
    # we should use the path given by the user to launch the client with subprocess.
    if DEFAULT_LEAGUE_CLIENT_EXECUTABLE in cli_args.launch_league:
        # If the default path is given, use the default launch arguments for league.
        commands = [
            cli_args.launch_league,
            "--launch-product=league_of_legends",
            "--launch-patchline=live",
        ]
    else:
        # If a custom path has been set, just execute that path
        commands = [cli_args.launch_league]

    subprocess.Popen(commands, shell=True)
