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
import inspect
import time
from dataclasses import dataclass, field
from threading import Thread, Timer

import pypresence
from lcu_driver.connection import Connection  # type:ignore

from league_rpc.lcu_api.helpers import (
    get_lcu_data_sync,
    handle_in_game,
    handle_spectating,
    show_ranked_data,
)
from league_rpc.models.client_data import ClientData
from league_rpc.models.lcu.current_chat_status import LolChatUser
from league_rpc.models.lcu.gameflow_phase import GameFlowPhase
from league_rpc.models.module_data import ModuleData
from league_rpc.models.rpc_data import RPCData
from league_rpc.utils.const import (
    BASE_MAP_ICON_URL,
    DEFAULT_MAP_ICON_FILENAME,
    GAME_MODE_CONVERT_MAP,
    LEAGUE_CLASSIC_ICON,
    MAP_ICON_CONVERT_MAP,
    MAP_ICON_FILENAME_OVERRIDES,
    PROFILE_ICON_BASE_URL,
    SMALL_TEXT,
)


# Discord has no publicly documented rate limit for activity updates. This value is a
# conservative, empirically-tuned guess, not a known threshold - adjust if testing shows
# presence updates stalling or clearing at this cadence.
HEARTBEAT_INTERVAL_SECONDS = 5

# League's own presence tends to react to the same state change we do, shortly after
# ours lands. A burst of resends right after a real update reclaims the display
# without waiting for the next periodic heartbeat.
RECLAIM_BURST_COUNT = 4
RECLAIM_BURST_INTERVAL_SECONDS = 1.5


# As some events are called multiple times, we should limit the amount of updates to the RPC.
# Collect update events for 1 second and then update the RPC.
@dataclass
class RPCUpdater:
    """A dataclass responsible for scheduling and executing updates to the Discord Rich Presence,
    encapsulating logic to delay and batch update requests to avoid rapid, unnecessary refreshes.
    """

    previous_client_data: ClientData | None = field(default=None, init=False)
    previous_rpc_data: RPCData | None = field(default=None, init=False)
    last_sent_was_clear: bool = field(default=False, init=False)
    last_sent_at: float = field(default=0.0, init=False)
    last_sent_details: str = field(default="", init=False)

    def trigger_rpc_update(
        self,
        module_data: ModuleData,
        clear_instead_of_update: bool = False,
    ) -> None:
        """
        Handles the update of the Rich Presence with the provided data, catching and logging any exceptions that occur.
        """

        # Debugging what function called trigger_rpc_update
        if module_data.cli_args.debug:  # type:ignore
            stack = inspect.stack()
            caller = stack[1].function
            module_data.logger.debug(f"Caller of trigger_rpc_update: {caller}")

        if self.has_rpc_data_changed(module_data.rpc_data):
            self.previous_rpc_data = copy.copy(module_data.rpc_data)
            try:
                module_data.logger.debug("Updating Discord Rich Presence")

                with module_data.rpc_lock:
                    if clear_instead_of_update:
                        module_data.logger.debug("Clearing Discord Rich Presence")
                        module_data.rpc.clear()  # type:ignore
                    else:
                        module_data.rpc.update(  # type: ignore
                            large_image=module_data.rpc_data.large_image,
                            large_text=module_data.rpc_data.large_text,
                            small_image=module_data.rpc_data.small_image,
                            small_text=module_data.rpc_data.small_text,
                            details=module_data.rpc_data.details,
                            state=module_data.rpc_data.state,
                            start=module_data.rpc_data.start,
                        )

                self.last_sent_was_clear = clear_instead_of_update
                self.last_sent_at = time.monotonic()
                self.last_sent_details = module_data.rpc_data.details

                if not clear_instead_of_update:
                    self._start_reclaim_burst(module_data)

            except Exception as e:
                module_data.logger.debug(
                    f"Exception occured while updating discord: {e}"
                )
        else:
            module_data.logger.debug("RPC data has not changed. Skipping update.")

    def start_heartbeat(self, module_data: ModuleData) -> None:
        """Starts a background loop that resends our last activity every
        HEARTBEAT_INTERVAL_SECONDS, so League's own Rich Presence can't outlast us.
        """
        Thread(target=self._heartbeat_loop, args=(module_data,), daemon=True).start()

    def _heartbeat_loop(self, module_data: ModuleData) -> None:
        # Sleep relative to the last actual send (real or heartbeat) rather than a
        # fixed schedule, so a real update mid-cycle doesn't cost a whole extra
        # HEARTBEAT_INTERVAL_SECONDS before the next heartbeat is allowed to fire.
        while True:
            time_since_last_send = time.monotonic() - self.last_sent_at
            time.sleep(max(HEARTBEAT_INTERVAL_SECONDS - time_since_last_send, 1))
            self._resend_last_activity(module_data)

    def _start_reclaim_burst(self, module_data: ModuleData) -> None:
        """Fires a few quick resends right after a real update, so we reclaim the
        display before League's own presence reacts to the same state change.
        """
        Thread(target=self._reclaim_burst, args=(module_data,), daemon=True).start()

    def _reclaim_burst(self, module_data: ModuleData) -> None:
        for _ in range(RECLAIM_BURST_COUNT):
            time.sleep(RECLAIM_BURST_INTERVAL_SECONDS)
            self._resend_last_activity(module_data, force=True)

    def _resend_last_activity(self, module_data: ModuleData, force: bool = False) -> None:
        """Re-sends the last activity we set, skipping if we have nothing to
        show, deliberately cleared our presence, or would exceed Discord's rate limit.
        """
        if self.previous_rpc_data is None or self.last_sent_was_clear:
            return
        if not force and time.monotonic() - self.last_sent_at < HEARTBEAT_INTERVAL_SECONDS:
            return

        # Discord appears to ignore a SET_ACTIVITY call if the activity payload is
        # byte-identical to what's already active, so a plain resend never displaces
        # League's own presence. Append a zero-width space, invisible but distinct,
        # whenever this would otherwise exactly repeat what we last transmitted.
        details = self.previous_rpc_data.details
        if details == self.last_sent_details:
            details += "​"

        try:
            with module_data.rpc_lock:
                module_data.rpc.update(  # type: ignore
                    large_image=self.previous_rpc_data.large_image,
                    large_text=self.previous_rpc_data.large_text,
                    small_image=self.previous_rpc_data.small_image,
                    small_text=self.previous_rpc_data.small_text,
                    details=details,
                    state=self.previous_rpc_data.state,
                    start=self.previous_rpc_data.start,
                )
            self.last_sent_at = time.monotonic()
            self.last_sent_details = details
            module_data.logger.debug(
                f"Heartbeat: resent activity at {time.strftime('%H:%M:%S')}"
            )
        except Exception as e:
            module_data.logger.debug(f"Heartbeat resend failed: {e}")

    def has_client_data_changed(self, current_client_data: ClientData) -> bool:
        """
        Compares the current client data with the previous client data to detect any changes.
        """
        if self.previous_client_data is None:
            # No previous data exists, so consider it as changed
            return True

        # Compare the current and previous client data
        return current_client_data != self.previous_client_data

    def has_rpc_data_changed(self, current_rpc_data: RPCData) -> bool:
        """
        Compares the current RPC data with the previous RPC data to detect any changes.
        """
        if self.previous_rpc_data is None:
            # No previous data exists, so consider it as changed
            return True

        # Compare the current and previous RPC data
        return current_rpc_data != self.previous_rpc_data

    def delay_update(
        self,
        module_data: ModuleData,
        connection: Connection,
    ) -> None:
        """Schedules an update if one is not already scheduled within a short delay (1 second)."""

        # Debugging what function called delay_update
        if module_data.cli_args.debug:  # type:ignore
            inspect.stack()
            caller = inspect.stack()[1].function
            module_data.logger.debug(f"Caller in delay_update: {caller}")

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
        )

    def in_client_rpc(
        self,
        module_data: ModuleData,
    ) -> None:
        """
        Updates Rich Presence when the user is in the client.
        """
        details: str = f"{module_data.client_data.availability}"
        hide_emojis: bool = module_data.cli_args.hide_emojis  # type:ignore
        clear_instead_of_update = False

        if not hide_emojis:
            status_emojis: str = f"{'🟢' if module_data.client_data.availability == LolChatUser.ONLINE.capitalize() else '  🔴'}"
            # details = status_emojis + details
            details = status_emojis + "  " + details

        module_data.rpc_data = RPCData(
            large_image=PROFILE_ICON_BASE_URL.format_map(
                {"icon_id": module_data.client_data.summoner_icon}
            ),
            large_text="In Client",
            small_image=LEAGUE_CLASSIC_ICON,
            small_text=SMALL_TEXT,
            details=details,
            state="In Client",
            start=module_data.client_data.application_start_time,
        )

        if module_data.cli_args and module_data.cli_args.hide_in_client:
            # If the user wants to hide the in-client RPC, we will clear it instead of updating it.
            clear_instead_of_update = True

        self.trigger_rpc_update(
            module_data,
            clear_instead_of_update,
        )

    def in_lobby_rpc(
        self,
        module_data: ModuleData,
    ) -> None:
        """Updates Rich Presence for lobby status, handling custom and standard lobbies."""
        large_image = PROFILE_ICON_BASE_URL.format_map(
            {"icon_id": module_data.client_data.summoner_icon}
        )

        large_text = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        small_image: str = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id, "classic_sru"),
            filename=MAP_ICON_FILENAME_OVERRIDES.get(
                module_data.client_data.map_id, DEFAULT_MAP_ICON_FILENAME
            ),
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
        if module_data.client_data.gamemode == "BRAWL":
            small_image = LEAGUE_CLASSIC_ICON
        if module_data.client_data.gamemode in ("JADE", "KIWI_JADE"):
            small_image = LEAGUE_CLASSIC_ICON

        module_data.rpc_data = RPCData(
            large_image=large_image,
            large_text=large_text,
            small_image=small_image,
            small_text=f"{small_text}",
            details=details,
            state=state,
            start=module_data.client_data.application_start_time,
        )
        self.trigger_rpc_update(module_data)

    def in_custom_lobby_rpc(
        self,
        module_data: ModuleData,
    ) -> None:
        """Updates Rich Presence for lobby status, handling custom and standard lobbies."""
        large_image = PROFILE_ICON_BASE_URL.format_map(
            {"icon_id": module_data.client_data.summoner_icon}
        )

        large_text = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        small_image: str = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id, "classic_sru"),
            filename=MAP_ICON_FILENAME_OVERRIDES.get(
                module_data.client_data.map_id, DEFAULT_MAP_ICON_FILENAME
            ),
        )
        small_text = SMALL_TEXT

        details: str = f"{module_data.client_data.queue_name}"
        state = "In Lobby"

        module_data.rpc_data = RPCData(
            large_image=large_image,
            large_text=large_text,
            small_image=small_image,
            small_text=small_text,
            details=details,
            state=state,
            start=module_data.client_data.application_start_time,
        )
        self.trigger_rpc_update(module_data)

    def in_queue_rpc(self, module_data: ModuleData) -> None:
        """Updates Rich Presence during the queue phase."""
        large_image: str = PROFILE_ICON_BASE_URL.format_map(
            {"icon_id": module_data.client_data.summoner_icon}
        )
        large_text: str = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        small_image: str = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id, "classic_sru"),
            filename=MAP_ICON_FILENAME_OVERRIDES.get(
                module_data.client_data.map_id, DEFAULT_MAP_ICON_FILENAME
            ),
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

        state = "In Queue"
        if (
            module_data.client_data.gameflow_phase
            == GameFlowPhase.CHECKED_INTO_TOURNAMENT
        ):
            state = "In Queue (Clash)"

        if module_data.client_data.gamemode == "BRAWL":
            small_image = LEAGUE_CLASSIC_ICON
        if module_data.client_data.gamemode in ("JADE", "KIWI_JADE"):
            small_image = LEAGUE_CLASSIC_ICON

        module_data.rpc_data = RPCData(
            large_image=large_image,
            large_text=large_text,
            small_image=small_image,
            small_text=small_text,
            details=f"{module_data.client_data.get_queue_name}",
            state=state,
            start=int(time.time()),
        )
        self.trigger_rpc_update(module_data)

    def in_champ_select_rpc(self, module_data: ModuleData) -> None:
        """Updates Rich Presence during champion selection."""
        large_image: str = PROFILE_ICON_BASE_URL.format_map(
            {"icon_id": module_data.client_data.summoner_icon}
        )
        large_text: str = f"{GAME_MODE_CONVERT_MAP.get(module_data.client_data.gamemode, module_data.client_data.gamemode)}"
        small_image: str = BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(module_data.client_data.map_id, "classic_sru"),
            filename=MAP_ICON_FILENAME_OVERRIDES.get(
                module_data.client_data.map_id, DEFAULT_MAP_ICON_FILENAME
            ),
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

        if module_data.client_data.gamemode == "BRAWL":
            small_image = LEAGUE_CLASSIC_ICON
        if module_data.client_data.gamemode in ("JADE", "KIWI_JADE"):
            small_image = LEAGUE_CLASSIC_ICON

        module_data.rpc_data = RPCData(
            large_image=large_image,
            large_text=large_text,
            small_image=small_image,
            small_text=small_text,
            details=f"{module_data.client_data.get_queue_name}",
            state="In Champ Select",
            start=int(time.time()),
        )
        self.trigger_rpc_update(module_data)

    # The function that updates discord rich presence, depending on the data
    def update_rpc(self, module_data: ModuleData, connection: Connection) -> None:
        """
        Determines the appropriate Rich Presence status based on the game flow phase and updates Discord.
        """
        data: ClientData = module_data.client_data
        rpc: pypresence.Presence | None = module_data.rpc

        if not isinstance(rpc, pypresence.Presence):
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
                self.in_client_rpc(module_data=module_data)
            case GameFlowPhase.WATCHING:
                handle_spectating(silent=False, module_data=module_data)
                while (
                    get_lcu_data_sync(
                        connection=connection,
                        endpoint="/lol-gameflow/v1/gameflow-phase",
                    )
                    == GameFlowPhase.WATCHING
                ):
                    handle_spectating(silent=True, module_data=module_data)
                    time.sleep(10)
                self.in_client_rpc(module_data=module_data)
            case GameFlowPhase.READY_CHECK:
                # When the READY check comes. We want to just ignore (IN_QUEUE rpc will still show.)
                return

            case (
                GameFlowPhase.NONE
                | GameFlowPhase.WAITING_FOR_STATS
                | GameFlowPhase.PRE_END_OF_GAME
                | GameFlowPhase.END_OF_GAME
            ):
                self.in_client_rpc(module_data=module_data)
                return
            case GameFlowPhase.CHAMP_SELECT | GameFlowPhase.GAME_START:
                # In Champ Select
                self.in_champ_select_rpc(module_data=module_data)
                return
            case (
                GameFlowPhase.MATCHMAKING
                | GameFlowPhase.READY_CHECK
                | GameFlowPhase.CHECKED_INTO_TOURNAMENT
            ):
                # In Queue
                self.in_queue_rpc(module_data=module_data)
                return
            case GameFlowPhase.LOBBY:
                # In Lobby
                if data.is_custom or data.is_practice:
                    self.in_custom_lobby_rpc(module_data=module_data)
                else:
                    # matchmaking lobby
                    self.in_lobby_rpc(module_data=module_data)
                return
            case GameFlowPhase.GAME_START:
                module_data.logger.info("Game is starting...")
            case GameFlowPhase.FAILED_TO_LAUNCH:
                module_data.logger.warning(
                    "Oops! League failed to launch. This issue comes from the game itself, not LeagueRPC. "
                    "If your game runs smoothly, feel free to ignore this. Otherwise, try restarting the client or checking for updates. "
                    "For persistent issues, reach out at https://discord.haze.sh for assistance."
                )
            case GameFlowPhase.RECONNECT:
                module_data.logger.info(
                    "Looks like you're reconnecting to your game. Hang in there and good luck!"
                )
            case GameFlowPhase.TERMINATED_IN_ERROR:
                module_data.logger.warning(
                    "The game has unexpectedly closed due to an error. This seems to be a League of Legends issue, not caused by LeagueRPC. "
                    "If everything else is running fine, you can safely ignore this message. Otherwise, consider restarting the client. "
                    "For persistent issues, reach out at https://discord.haze.sh for assistance."
                )
            case _:
                # other unhandled gameflow phases
                module_data.logger.warning(
                    f"Unhandled Gameflow Phase: {data.gameflow_phase}"
                )
