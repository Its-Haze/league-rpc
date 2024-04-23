"""
This module holds the Colors class and related functionalities around it.
"""

from dataclasses import dataclass

from league_rpc.__version__ import (
    RELEASES_PAGE,
    __version__,
    check_latest_version,
    get_version_from_github,
)


@dataclass
class Color:
    """
    Dataclass, storing the different colors that is used in the program.
    """

    dred: str = "\033[31m"
    dgreen: str = "\033[32m"
    yellow: str = "\033[33m"
    dblue: str = "\033[34m"
    dmagenta: str = "\033[35m"
    dcyan: str = "\033[36m"
    lgrey: str = "\033[37m"
    dgray: str = "\033[90m"
    red: str = "\033[91m"
    green: str = "\033[92m"
    orange: str = "\033[93m"
    blue: str = "\033[94m"
    magenta: str = "\033[95m"
    cyan: str = "\033[96m"
    white: str = "\033[97m"
    reset: str = "\033[0m"

    @property
    def logo(self) -> str:
        """Just prints the LEAGUE rpc logo, in your favorite Terminal Emulator."""

        return rf"""
        {self.yellow}  _                                  {self.dblue} _____  _____   _____ {self.reset}
        {self.yellow} | |                                 {self.dblue}|  __ \|  __ \ / ____|{self.reset}
        {self.yellow} | |     ___  __ _  __ _ _   _  ___  {self.dblue}| |__) | |__) | |     {self.reset}
        {self.yellow} | |    / _ \/ _` |/ _` | | | |/ _ \ {self.dblue}|  _  /|  ___/| |     {self.reset}
        {self.yellow} | |___|  __/ (_| | (_| | |_| |  __/ {self.dblue}| | \ \| |    | |____ {self.reset}
        {self.yellow} |______\___|\__,_|\__, |\__,_|\___| {self.dblue}|_|  \_\_|     \_____|{self.reset}
        {self.yellow}                    __/ |                                              {self.reset}
        {self.yellow}                   |___/ By @Haze.dev - (Version: {self.green if not check_latest_version() else self.red}{__version__}{self.yellow})       {self.reset}
        {f'{self.yellow} A newer version is available at {RELEASES_PAGE} {self.green}[{get_version_from_github()}]{self.reset}' if check_latest_version() else ""}
        """
