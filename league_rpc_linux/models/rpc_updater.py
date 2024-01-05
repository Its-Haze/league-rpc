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
from league_rpc_linux.lcu_api.lcu_connector import ModuleData
from league_rpc_linux.models.lcu.current_chat_status import LolChatUser
from league_rpc_linux.models.lcu.gameflow_phase import GameFlowPhase


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

    @staticmethod
    def in_client_rpc(rpc: Presence, module_data: ModuleData) -> None:
        """
        Updates the RPC when the use is in the client.
        """
        state = f"{module_data.client_data.availability}"
        show_emojis: bool = module_data.cli_args.show_emojis  # type:ignore

        if show_emojis:
            status_emojis = f"{'🟢' if module_data.client_data.availability == LolChatUser.ONLINE.capitalize() else' 🔴'}"
            state = status_emojis + state

        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{module_data.client_data.summoner_icon}.png",
            large_text="In Client",
            small_image=LEAGUE_OF_LEGENDS_LOGO,
            small_text=SMALL_TEXT,
            details="In Client",
            state=state,
        )

    @staticmethod
    def in_lobby_rpc(rpc: Presence, module_data: ModuleData, is_custom: bool) -> None:
        if is_custom:
            large_image = f"{PROFILE_ICON_BASE_URL}{str(module_data.client_data.summoner_icon)}.png"

            large_text = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
            small_image = BASE_MAP_ICON_URL.format(
                map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id)
            )
            small_text = SMALL_TEXT

            details = f"In Lobby: {module_data.client_data.queue}"
            state = "Custom Lobby"

            rpc.update(
                large_image=large_image,
                large_text=large_text,
                small_image=small_image,
                small_text=small_text,
                details=details,
                state=state,
            )
        else:
            large_image = f"{PROFILE_ICON_BASE_URL}{str(module_data.client_data.summoner_icon)}.png"

            large_text = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"

            small_image = BASE_MAP_ICON_URL.format(
                map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id)
            )
            small_text = SMALL_TEXT
            details = f"{module_data.client_data.queue}"
            state = f"In Lobby ({module_data.client_data.players}/{module_data.client_data.max_players})"

            if module_data.cli_args.show_rank:
                match module_data.client_data.queue:
                    case "Ranked Solo/Duo":
                        summoner_rank = module_data.client_data.summoner_rank
                        if summoner_rank.tier:
                            (
                                small_text,
                                small_image,
                            ) = summoner_rank.rpc_info
                            large_text = SMALL_TEXT

                    case "Ranked Flex":
                        summoner_rank = module_data.client_data.summoner_rank_flex
                        if summoner_rank.tier:
                            (
                                small_text,
                                small_image,
                            ) = summoner_rank.rpc_info
                            large_text = SMALL_TEXT

                    case "Teamfight Tactics (Ranked)":
                        summoner_rank = module_data.client_data.tft_rank
                        if summoner_rank.tier:
                            (
                                small_text,
                                small_image,
                            ) = summoner_rank.rpc_info
                            large_text = SMALL_TEXT
                    case "Arena":
                        summoner_rank = module_data.client_data.arena_rank
                        if summoner_rank.tier:
                            (
                                small_text,
                                small_image,
                            ) = summoner_rank.rpc_info
                            large_text = SMALL_TEXT

                    case _:
                        ...

            rpc.update(
                large_image=large_image,
                large_text=large_text,
                small_image=small_image,
                small_text=f"{small_text}",
                details=details,
                state=state,
            )

    @staticmethod
    def in_queue_rpc(rpc: Presence, module_data: ModuleData) -> None:
        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{module_data.client_data.summoner_icon}.png",
            large_text=f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}",
            small_image=BASE_MAP_ICON_URL.format(
                map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id)
            )
            or LEAGUE_OF_LEGENDS_LOGO,
            small_text=SMALL_TEXT,
            details=f"{module_data.client_data.queue}",
            state="In Queue",
            start=int(time.time()),
        )


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
            # Handled by the "inGame" flow in __main__.py
            return
        case GameFlowPhase.READY_CHECK:
            # When the READY check comes. We want to just ignore (IN_QUEUE rpc will still show.)
            return

        case GameFlowPhase.NONE | GameFlowPhase.WAITING_FOR_STATS | GameFlowPhase.PRE_END_OF_GAME | GameFlowPhase.END_OF_GAME:
            RPCUpdater.in_client_rpc(rpc, module_data)
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
                details=f"In Champ Select: {data.queue} - ({data.players}/{data.max_players})",
                state="Picking Champions...",
                party_size=[data.players, data.max_players],
            )
            return
        case GameFlowPhase.MATCHMAKING | GameFlowPhase.READY_CHECK:
            # In Queue
            RPCUpdater.in_queue_rpc(rpc, module_data)
            return
        case GameFlowPhase.LOBBY:
            # In Lobby
            if data.is_custom or data.is_practice:
                RPCUpdater.in_lobby_rpc(rpc, module_data, is_custom=True)
            else:
                # matchmaking lobby
                # TODO: Add checks for which queue/gamemode type the lobby is.. to show correct rank info.
                RPCUpdater.in_lobby_rpc(rpc, module_data, is_custom=False)
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
