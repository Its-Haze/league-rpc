import sys
import time

import pypresence

from league_rpc_linux.champion import gather_ingame_information, get_skin_asset
from league_rpc_linux.colors import Colors
from league_rpc_linux.gametime import get_current_ingame_time
from league_rpc_linux.kda import get_kda
from league_rpc_linux.processes.process import (
    check_discord_process,
    check_league_client_process,
    check_riot_games_service_process,
    player_state,
)


def main():
    """
    main program.
    """
    # Initialize
    print(Colors().logo)

    discord_process_names = ["Discord", "DiscordPTB", "DiscordCanary", "electron"]
    rpc = check_discord_process(discord_process_names)

    check_riot_games_service_process()
    check_league_client_process()

    print(f"{Colors.green}\nRich Presence utilized!{Colors.reset}")
    league_of_legends_logo = "https://freepngimg.com/save/85643-blue-league-legends-icons-of-symbol-garena/1600x1600.png"
    start_time = int(time.time())

    while True:
        try:
            match player_state():
                case "InGame":
                    print(
                        f"{Colors.dblue}Detected game! Will soon gather data and update discord RPC{Colors.reset}"
                    )

                    (
                        champ_name,
                        skin_id,
                        gamemode,
                        _,
                    ) = gather_ingame_information()
                    if gamemode != "TFT":
                        # LEAGUE RPC
                        skin_asset = get_skin_asset(
                            champion_name=champ_name, skin_id=skin_id
                        )
                        while player_state() == "InGame":
                            rpc.update(  # type:ignore
                                large_image=skin_asset,
                                large_text=champ_name,
                                details=gamemode,
                                state=f"In Game {get_kda()}",
                                small_image=league_of_legends_logo,
                                small_text="github.com/Its-Haze/league-rpc-linux",
                                start=int(time.time())
                                - get_current_ingame_time(default_time=start_time),
                            )
                            if champ_name == "???" or gamemode == "???":
                                print(
                                    f"{Colors.red}Failed to load in data.. {Colors.lgrey}will try again shortly.\n{Colors.dcyan}(Reason: Someone has potato PC){Colors.reset}"
                                )
                                break
                            time.sleep(10)
                    else:
                        # TFT RPC
                        while player_state() == "InGame":
                            (
                                _,
                                _,
                                gamemode,
                                level,
                            ) = gather_ingame_information()
                            if gamemode and level:
                                rpc.update(  # type:ignore
                                    large_image="https://wallpapercave.com/wp/wp7413493.jpg",
                                    large_text="Playing TFT",
                                    details="Teamfight Tactics",
                                    state=f"In Game · lvl: {level}",
                                    small_image=league_of_legends_logo,
                                    small_text="github.com/Its-Haze/league-rpc-linux",
                                    start=int(time.time())
                                    - get_current_ingame_time(default_time=start_time),
                                )
                            time.sleep(10)

                case "InLobby":
                    rpc.update(  # type:ignore
                        large_image=league_of_legends_logo,
                        large_text="github.com/Its-Haze/league-rpc-linux",
                        state="In Client",
                        start=start_time,
                    )
                    time.sleep(10)

                case _:
                    print(
                        f"{Colors.red}LeagueOfLegends.exe was terminated. rpc shuting down..{Colors.reset}."
                    )
                    sys.exit()
        except pypresence.exceptions.PipeClosed:
            print(
                f"{Colors.red} Discord seems to be closed. Reconnect, and restart this script. {Colors.reset}"
            )
            sys.exit()


if __name__ == "__main__":
    main()
