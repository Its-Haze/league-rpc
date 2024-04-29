import subprocess
import time
from argparse import Namespace

from league_rpc.utils.const import DEFAULT_LEAGUE_CLIENT_EXECUTABLE


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
    time.sleep(5)
