"""
This module provides the RPCUpdater dataclass designed to manage and throttle updates to Discord Rich Presence
based on events occurring within the League of Legends client. The updater focuses on limiting redundant updates
to optimize performance and ensure that the displayed status is current without overwhelming the update mechanism.

Usage:
    RPCUpdater is critical in applications where live updates of user status are necessary, especially in contexts where
    the user's game state changes frequently, such as during matchmaking or in various phases of gameplay. It ensures
    that updates are efficiently managed and that the displayed information remains accurate without excessive updates,
    which could disrupt the user experience or exceed API rate limits.
"""

import copy
import time
from dataclasses import dataclass, field
from threading import Timer

from lcu_driver.connection import Connection  # type:ignore
from pypresence import Presence  # type:ignore

from league_rpc.lcu_api.helpers import (
    get_lcu_data_sync,
    handle_in_game,
    show_ranked_data,
)
from league_rpc.models.client_data import ClientData
from league_rpc.models.lcu.current_chat_status import LolChatUser
from league_rpc.models.lcu.gameflow_phase import GameFlowPhase
from league_rpc.models.module_data import ModuleData
from league_rpc.utils.const import (
    BASE_MAP_ICON_URL,
    GAME_MODE_CONVERT_MAP,
    LEAGUE_OF_LEGENDS_LOGO,
    MAP_ICON_CONVERT_MAP,
    PROFILE_ICON_BASE_URL,
    SMALL_TEXT,
)


# As some events are called multiple times, we should limit the amount of updates to the RPC.
# Collect update events for 1 second and then update the RPC.
@dataclass
class RPCUpdater:
    """A dataclass responsible for scheduling and executing updates to the Discord Rich Presence,
    encapsulating logic to delay and batch update requests to avoid rapid, unnecessary refreshes.
    """

    previous_client_data: ClientData | None = field(default=None, init=False)

    def has_client_data_changed(self, current_client_data: ClientData) -> bool:
        """
        Compares the current client data with the previous client data to detect any changes.
        """
        if self.previous_client_data is None:
            # No previous data exists, so consider it as changed
            return True

        # Compare the current and previous client data
        return current_client_data != self.previous_client_data

    def delay_update(
        self,
        module_data: ModuleData,
        connection: Connection,
    ) -> None:
        """Schedules an update if one is not already scheduled within a short delay (1 second)."""
        # Check if the client data has changed
        if self.has_client_data_changed(module_data.client_data):
            Timer(
                interval=1.5,
                function=self.update_rpc_and_reset_flag,
                args=(module_data, connection),
            ).start()

    def update_rpc_and_reset_flag(
        self, module_data: ModuleData, connection: Connection
    ) -> None:
        """Executes the update to Rich Presence and resets the scheduling flag."""
        # Store the current client data as the previous state
        self.previous_client_data = copy.copy(module_data.client_data)

        self.update_rpc(
            module_data=module_data,
            connection=connection,
        )  # Assuming update_rpc is defined elsewhere

    @staticmethod
    def in_client_rpc(
        rpc: Presence,
        module_data: ModuleData,
    ) -> None:
        """
        Updates Rich Presence when the user is in the client.
        """
        details: str = f"{module_data.client_data.availability}"
        hide_emojis: bool = module_data.cli_args.hide_emojis  # type:ignore

        if not hide_emojis:
            status_emojis: str = (
                f"{'🟢' if module_data.client_data.availability == LolChatUser.ONLINE.capitalize() else '  🔴'}"
            )
            # details = status_emojis + details
            details = status_emojis + "  " + details

        try:
            rpc.update(  # type: ignore
                large_image=f"{PROFILE_ICON_BASE_URL}{module_data.client_data.summoner_icon}.png",
                large_text="In Client",
                small_image=LEAGUE_OF_LEGENDS_LOGO,
                small_text=SMALL_TEXT,
                details=details,
                state="In Client",
                start=module_data.client_data.application_start_time,
            )
        except RuntimeError:
            module_data.logger.debug("Error in Client RPC: Probably safe to ignore")

    @staticmethod
    def in_lobby_rpc(
        rpc: Presence,
        module_data: ModuleData,
    ) -> None:
        """Updates Rich Presence for lobby status, handling custom and standard lobbies."""
        large_image = (
            f"{PROFILE_ICON_BASE_URL}{str(module_data.client_data.summoner_icon)}.png"
        )

        large_text = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        small_image: str = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id)
        )
        small_text = SMALL_TEXT

        details = f"{module_data.client_data.get_queue_name}"
        state = f"In Lobby ({module_data.client_data.players}/{module_data.client_data.max_players})"

        if not module_data.cli_args.no_rank:  # type: ignore
            _large_text, _small_image, _small_text = show_ranked_data(module_data)
            if all([_large_text, _small_image, _small_text]):
                large_text, small_image, small_text = (
                    _large_text,
                    _small_image,
                    _small_text,
                )
        if module_data.client_data.gamemode == "TFT":
            large_image = module_data.client_data.tft_companion_icon
            large_text = module_data.client_data.tft_companion_name

        try:
            rpc.update(  # type: ignore
                large_image=large_image,
                large_text=large_text,
                small_image=small_image,
                small_text=f"{small_text}",
                details=details,
                state=state,
                start=module_data.client_data.application_start_time,
            )
        except RuntimeError:
            module_data.logger.debug("Error in Lobby RPC: Probably safe to ignore")

    @staticmethod
    def in_custom_lobby_rpc(
        rpc: Presence,
        module_data: ModuleData,
    ) -> None:
        """Updates Rich Presence for lobby status, handling custom and standard lobbies."""
        large_image = (
            f"{PROFILE_ICON_BASE_URL}{str(module_data.client_data.summoner_icon)}.png"
        )

        large_text = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        small_image: str = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id)
        )
        small_text = SMALL_TEXT

        details: str = f"In Lobby: {module_data.client_data.queue_name}"
        state = "Custom Lobby"
        try:
            rpc.update(  # type: ignore
                large_image=large_image,
                large_text=large_text,
                small_image=small_image,
                small_text=small_text,
                details=details,
                state=state,
                start=module_data.client_data.application_start_time,
            )
        except RuntimeError:
            module_data.logger.debug(
                "Error in Custom Lobby RPC: Probably safe to ignore"
            )

    @staticmethod
    def in_queue_rpc(rpc: Presence, module_data: ModuleData) -> None:
        """Updates Rich Presence during the queue phase."""
        large_image: str = (
            f"{PROFILE_ICON_BASE_URL}{module_data.client_data.summoner_icon}.png"
        )
        large_text: str = (
            f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        )
        small_image: str = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id)
        )
        small_text = SMALL_TEXT

        if not module_data.cli_args.no_rank:  # type: ignore
            _large_text, _small_image, _small_text = show_ranked_data(module_data)
            if all([_large_text, _small_image, _small_text]):
                large_text, small_image, small_text = (
                    _large_text,
                    _small_image,
                    _small_text,
                )
        try:
            rpc.update(  # type: ignore
                large_image=large_image,
                large_text=large_text,
                small_image=small_image,
                small_text=small_text,
                details=f"{module_data.client_data.get_queue_name}",
                state="In Queue",
                start=int(time.time()),
            )
        except RuntimeError:
            module_data.logger.debug("Error in Queue RPC: Probably safe to ignore")

    @staticmethod
    def in_champ_select_rpc(rpc: Presence, module_data: ModuleData) -> None:
        """Updates Rich Presence during champion selection."""
        large_image: str = (
            f"{PROFILE_ICON_BASE_URL}{module_data.client_data.summoner_icon}.png"
        )
        large_text: str = (
            f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        )
        small_image: str = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id)
        )
        small_text = SMALL_TEXT

        if not module_data.cli_args.no_rank:  # type: ignore
            _large_text, _small_image, _small_text = show_ranked_data(module_data)
            if all([_large_text, _small_image, _small_text]):
                large_text, small_image, small_text = (
                    _large_text,
                    _small_image,
                    _small_text,
                )
        try:
            rpc.update(  # type: ignore
                large_image=large_image,
                large_text=large_text,
                small_image=small_image,
                small_text=small_text,
                details=f"{module_data.client_data.get_queue_name}",
                state="In Champ Select",
                start=int(time.time()),
            )
        except RuntimeError:
            module_data.logger.debug(
                "Error in Champ Select RPC: Probably safe to ignore"
            )

    # The function that updates discord rich presence, depending on the data
    def update_rpc(self, module_data: ModuleData, connection: Connection) -> None:
        """
        Determines the appropriate Rich Presence status based on the game flow phase and updates Discord.
        """
        data: ClientData = module_data.client_data
        rpc: Presence | None = module_data.rpc

        if not isinstance(rpc, Presence):
            # Only continue if rpc is of type Presence.
            module_data.logger.error("RPC is not of type Presence")
            return

        match data.gameflow_phase:
            # This value will be set by "/lol-gameflow/v1/gameflow-phase"
            case GameFlowPhase.IN_PROGRESS:
                handle_in_game(
                    connection=connection,
                    silent=False,
                    module_data=module_data,
                )  # Print champion details
                while (
                    get_lcu_data_sync(
                        connection=connection,
                        endpoint="/lol-gameflow/v1/gameflow-phase",
                    )
                    == GameFlowPhase.IN_PROGRESS
                ):
                    handle_in_game(
                        connection=connection,
                        silent=True,  # No prints here, since we've already done so, just update the RPC
                        module_data=module_data,
                    )
                    time.sleep(10)
                # After the game is over, we will drop back to the main client.
                self.in_client_rpc(rpc=rpc, module_data=module_data)
            case GameFlowPhase.READY_CHECK:
                # When the READY check comes. We want to just ignore (IN_QUEUE rpc will still show.)
                return

            case (
                GameFlowPhase.NONE
                | GameFlowPhase.WAITING_FOR_STATS
                | GameFlowPhase.PRE_END_OF_GAME
                | GameFlowPhase.END_OF_GAME
            ):
                self.in_client_rpc(rpc=rpc, module_data=module_data)
                return
            case GameFlowPhase.CHAMP_SELECT | GameFlowPhase.GAME_START:
                # In Champ Select
                self.in_champ_select_rpc(rpc=rpc, module_data=module_data)
                return
            case GameFlowPhase.MATCHMAKING | GameFlowPhase.READY_CHECK:
                # In Queue
                self.in_queue_rpc(rpc=rpc, module_data=module_data)
                return
            case GameFlowPhase.LOBBY:
                # In Lobby
                if data.is_custom or data.is_practice:
                    self.in_custom_lobby_rpc(rpc=rpc, module_data=module_data)
                else:
                    # matchmaking lobby
                    self.in_lobby_rpc(rpc=rpc, module_data=module_data)
                return
            case GameFlowPhase.GAME_START:
                module_data.logger.info("Game is starting...")
            case _:
                # other unhandled gameflow phases
                module_data.logger.warning(
                    f"Unhandled Gameflow Phase: {data.gameflow_phase}"
                )
