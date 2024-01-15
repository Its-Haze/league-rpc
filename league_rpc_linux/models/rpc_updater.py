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
    RANKED_TYPE_MAPPER,
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
        self.update_rpc(module_data)  # Assuming update_rpc is defined elsewhere
        self.scheduled_update = False

    @staticmethod
    def in_client_rpc(
        rpc: Presence,
        module_data: ModuleData,
    ) -> None:
        """
        Updates the RPC when the use is in the client.
        """
        details = f"{module_data.client_data.availability}"
        show_emojis: bool = module_data.cli_args.show_emojis  # type:ignore

        if show_emojis:
            status_emojis = f"{'🟢' if module_data.client_data.availability == LolChatUser.ONLINE.capitalize() else '  🔴'}"
            # details = status_emojis + details
            details = status_emojis + "  " + details

        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{module_data.client_data.summoner_icon}.png",
            large_text="In Client",
            small_image=LEAGUE_OF_LEGENDS_LOGO,
            small_text=SMALL_TEXT,
            details=details,
            state="In Client",
            start=module_data.client_data.application_start_time,
        )

    @staticmethod
    def in_lobby_rpc(
        rpc: Presence,
        module_data: ModuleData,
        is_custom: bool,
    ) -> None:
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
                start=module_data.client_data.application_start_time,
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
                _large_text, _small_image, _small_text = RPCUpdater.show_ranked_data(
                    module_data
                )
                if all([_large_text, _small_image, _small_text]):
                    large_text, small_image, small_text = (
                        _large_text,
                        _small_image,
                        _small_text,
                    )

            rpc.update(
                large_image=large_image,
                large_text=large_text,
                small_image=small_image,
                small_text=f"{small_text}",
                details=details,
                state=state,
                start=module_data.client_data.application_start_time,
            )

    @staticmethod
    def show_ranked_data(
        module_data: ModuleData,
    ) -> tuple[str, ...]:
        large_text = small_text = small_image = ""

        match RANKED_TYPE_MAPPER.get(module_data.client_data.queue_type):
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
        return large_text, small_image, small_text

    @staticmethod
    def in_queue_rpc(rpc: Presence, module_data: ModuleData) -> None:
        large_image = (
            f"{PROFILE_ICON_BASE_URL}{module_data.client_data.summoner_icon}.png"
        )
        large_text = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        small_image = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id)
        )
        small_text = SMALL_TEXT

        if module_data.cli_args.show_rank:
            _large_text, _small_image, _small_text = RPCUpdater.show_ranked_data(
                module_data
            )
            if all([_large_text, _small_image, _small_text]):
                large_text, small_image, small_text = (
                    _large_text,
                    _small_image,
                    _small_text,
                )

        rpc.update(
            large_image=large_image,
            large_text=large_text,
            small_image=small_image,
            small_text=small_text,
            details=f"{module_data.client_data.queue}",
            state="In Queue",
            start=int(time.time()),
        )

    @staticmethod
    def in_champ_select_rpc(rpc: Presence, module_data: ModuleData) -> None:
        large_image = (
            f"{PROFILE_ICON_BASE_URL}{module_data.client_data.summoner_icon}.png"
        )
        large_text = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        small_image = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id)
        )
        small_text = SMALL_TEXT

        if module_data.cli_args.show_rank:
            _large_text, _small_image, _small_text = RPCUpdater.show_ranked_data(
                module_data
            )
            if all([_large_text, _small_image, _small_text]):
                large_text, small_image, small_text = (
                    _large_text,
                    _small_image,
                    _small_text,
                )

        rpc.update(
            large_image=large_image,
            large_text=large_text,
            small_image=small_image,
            small_text=small_text,
            details=f"{module_data.client_data.queue}",
            state="In Champ Select",
            start=int(time.time()),
        )

    # The function that updates discord rich presence, depending on the data
    def update_rpc(self, module_data: ModuleData):
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
                self.in_client_rpc(rpc, module_data)
                return
            case GameFlowPhase.CHAMP_SELECT | GameFlowPhase.GAME_START:
                # In Champ Select
                self.in_champ_select_rpc(rpc, module_data)
                return
            case GameFlowPhase.MATCHMAKING | GameFlowPhase.READY_CHECK:
                # In Queue
                self.in_queue_rpc(rpc, module_data)
                return
            case GameFlowPhase.LOBBY:
                # In Lobby
                if data.is_custom or data.is_practice:
                    self.in_lobby_rpc(rpc, module_data, is_custom=True)
                else:
                    # matchmaking lobby
                    self.in_lobby_rpc(rpc, module_data, is_custom=False)
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
                    start=module_data.client_data.application_start_time,
                )
