import time
from dataclasses import dataclass
from threading import Timer

from pypresence import Presence

from league_rpc_linux.const import (
    BASE_MAP_ICON_URL,
    GAME_MODE_CONVERT_MAP,
    LEAGUE_OF_LEGENDS_LOGO,
    MAP_ICON_CONVERT_MAP,
    PROFILE_ICON_BASE_URL,
    SMALL_TEXT,
)
from league_rpc_linux.lcu_api.gameflow_phase import GameFlowPhase
from league_rpc_linux.processes.lcu_thread import ModuleData


# As some events are called multiple times, we should limit the amount of updates to the RPC.
# Collect update events for 1 second and then update the RPC.
@dataclass
class RPCUpdater:
    scheduled_update: bool = False

    def delay_update(self, module_data: ModuleData):
        if not self.scheduled_update:
            self.scheduled_update = True
            Timer(1.0, self.update_rpc_and_reset_flag, args=(module_data,)).start()

    def update_rpc_and_reset_flag(self, module_data: ModuleData):
        update_rpc(module_data)  # Assuming update_rpc is defined elsewhere
        self.scheduled_update = False


# The function that updates discord rich presence, depending on the data
def update_rpc(module_data: ModuleData):
    print("Updating Discord Presence.")
    data = module_data.client_data
    rpc = module_data.rpc

    if not isinstance(rpc, Presence):
        # Only continue if rpc is of type Presence.
        return

    match data.gameflow_phase:
        # This value will be set by "/lol-gameflow/v1/gameflow-phase"

        case GameFlowPhase.IN_PROGRESS:
            print("DID YOU JUST ENTER A GAME?? :O")
            return

        case GameFlowPhase.NONE | GameFlowPhase.WAITING_FOR_STATS | GameFlowPhase.PRE_END_OF_GAME | GameFlowPhase.END_OF_GAME:
            rpc.update(
                large_image=f"{PROFILE_ICON_BASE_URL}{data.summoner_icon}.png",
                large_text="In Client",
                small_image=LEAGUE_OF_LEGENDS_LOGO,
                small_text=SMALL_TEXT,
                details="In Client",
                state=f"{data.availability}",
            )
            return
        case GameFlowPhase.CHAMP_SELECT:
            # In Champ Select

            rpc.update(
                large_image=f"{PROFILE_ICON_BASE_URL}{data.summoner_icon}.png",
                large_text=f"{GAME_MODE_CONVERT_MAP.get(data.gamemode, data.gamemode)}",
                small_image=BASE_MAP_ICON_URL.format(
                    map_name=MAP_ICON_CONVERT_MAP.get(data.map_id)
                )
                or LEAGUE_OF_LEGENDS_LOGO,
                small_text=SMALL_TEXT,
                details=f"In Champ Select: {data.queue}",
                state="Picking Champions...",
                party_size=[data.players, data.max_players],
            )
            return
        case GameFlowPhase.MATCHMAKING:
            # In Queue
            rpc.update(
                large_image=f"{PROFILE_ICON_BASE_URL}{data.summoner_icon}.png",
                large_text=f"{GAME_MODE_CONVERT_MAP.get(data.gamemode, data.gamemode)}",
                small_image=BASE_MAP_ICON_URL.format(
                    map_name=MAP_ICON_CONVERT_MAP.get(data.map_id)
                )
                or LEAGUE_OF_LEGENDS_LOGO,
                small_text=SMALL_TEXT,
                details=f"In Queue: {data.queue}",
                state="Searching for Game...",
                start=int(time.time()),
            )
            return
        case GameFlowPhase.LOBBY:
            # In Lobby
            if data.is_custom or data.is_practice:
                # custom or practice tool lobby
                large_image = f"{PROFILE_ICON_BASE_URL}{str(data.summoner_icon)}.png"
                large_text = (
                    f"{GAME_MODE_CONVERT_MAP.get(data.gamemode, data.gamemode)}"
                )
                small_image = BASE_MAP_ICON_URL.format(
                    map_name=MAP_ICON_CONVERT_MAP.get(data.map_id)
                )
                small_text = SMALL_TEXT
                details = f"In Lobby: {data.queue}"
                state = "Custom Lobby"
                print("--------")
                print(large_image)
                print(large_text)
                print(small_image)
                print(small_text)
                print(details)
                print(state)
                print("--------")

                rpc.update(
                    large_image=large_image,
                    large_text=large_text,
                    small_image=small_image,
                    small_text=small_text,
                    details=details,
                    state=state,
                )

            else:
                # matchmaking lobby
                rpc.update(
                    large_image=f"{PROFILE_ICON_BASE_URL}{str(data.summoner_icon)}.png",
                    large_text=f"{GAME_MODE_CONVERT_MAP.get(data.gamemode, data.gamemode)}",
                    small_image=BASE_MAP_ICON_URL.format(
                        map_name=MAP_ICON_CONVERT_MAP.get(data.map_id)
                    )
                    or LEAGUE_OF_LEGENDS_LOGO,
                    small_text=SMALL_TEXT,
                    details=f"In Lobby: {data.queue}",
                    party_size=[data.players, data.max_players],
                    state="Waiting for Players...",
                )
            return
        case _:
            # other unhandled gameflow phases
            print(f"Unhandled Gameflow Phase: {data.gameflow_phase}")
            rpc.update(
                large_image=f"{PROFILE_ICON_BASE_URL}{str(data.summoner_icon)}.png",
                large_text=f"{data.gameflow_phase}",
                small_image=LEAGUE_OF_LEGENDS_LOGO,
                small_text=SMALL_TEXT,
                details=f"{data.gameflow_phase}",
                state="Unhandled Gameflow Phase",
            )
