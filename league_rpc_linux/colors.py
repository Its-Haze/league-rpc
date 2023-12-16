"""
This module holds the Colors class and related functionalities around it.
"""

from dataclasses import dataclass


@dataclass
class Colors:
    """
    Dataclass, storing the different colors that is used in the program.
    """

    dred = "\033[31m"
    dgreen = "\033[32m"
    yellow = "\033[33m"
    dblue = "\033[34m"
    dmagenta = "\033[35m"
    dcyan = "\033[36m"
    lgrey = "\033[37m"
    dgray = "\033[90m"
    red = "\033[91m"
    green = "\033[92m"
    orange = "\033[93m"
    blue = "\033[94m"
    magenta = "\033[95m"
    cyan = "\033[96m"
    white = "\033[97m"
    reset = "\033[0m"

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
        {self.yellow}                    __/ |                                                {self.reset}
        {self.yellow}                   |___/                                                 {self.reset}
        """
