import time
from argparse import Namespace
from typing import Any, Optional

from aiohttp import ClientResponse
from lcu_driver.connection import Connection  # type:ignore
from lcu_driver.events.responses import WebsocketEventResponse  # type:ignore
from pypresence import Presence  # type:ignore

from league_rpc.disable_native_rpc.disable import (
    check_plugin_status,
    find_game_path,
    add_plugin,
    DISCORD_PLUGIN_BLOB,
)
from league_rpc.lcu_api.base_data import gather_base_data, set_tft_companion_data
from league_rpc.logger.richlogger import RichLogger
from league_rpc.models.client_data import ArenaStats, ClientData, RankedStats, TFTStats
from league_rpc.models.lcu.current_chat_status import LolChatUser
from league_rpc.models.lcu.current_summoner import Summoner
from league_rpc.models.lcu.gameflow_phase import GameFlowPhase
from league_rpc.models.module_data import ModuleData
from league_rpc.models.rpc_data import RPCData
from league_rpc.models.rpc_updater import RPCUpdater
from league_rpc.processes.process import processes_exists

module_data = ModuleData(
    client_data=ClientData(),
    rpc_updater=RPCUpdater(),
    rpc_data=RPCData(),
)


@module_data.connector.ready  # type:ignore
async def connect(connection: Connection) -> None:
    """
    This function will be called when the connection to the League Client API is established.
    """
    logger = module_data.logger
    rpc_updater = module_data.rpc_updater

    logger.start_progress_bar(name="Start LeagueRPC Engine")
    time.sleep(1)

    logger.info("Connected to the League Client API.")
    logger.update_progress_bar(advance=20)

    time.sleep(0.5)  # Give the client some time to load
    logger.update_progress_bar(advance=10)

    time.sleep(0.5)  # Give the client some time to load
    logger.update_progress_bar(advance=10)

    time.sleep(0.5)
    logger.update_progress_bar(advance=20)

    time.sleep(2)
    await gather_base_data(connection=connection, module_data=module_data)
    logger.info("Successfully gathered base data.")
    logger.update_progress_bar(advance=30)

    rpc_updater.delay_update(module_data=module_data, connection=connection)
    logger.info("Discord RPC successfully updated")
    logger.update_progress_bar(advance=40)

    logger.stop_progress_bar()

    time.sleep(0.5)

    logger.info("LeagueRPC is ready!", color="cyan")

    if game_path := find_game_path():
        native_presence = check_plugin_status(file_path=game_path, logger=logger)
        if native_presence:
            logger.warning(
                "The Native League Presence is still active. Please start this application before launching League of legends to fully disable it.",
                color="yellow",
            )


@module_data.connector.close  # type:ignore
async def disconnect(_: Connection) -> None:
    """
    This function will be called when the connection to the League Client API is closed.
    """
    logger = module_data.logger
    logger.info("Disconnected from the League Client API.", color="red")

    league_processes: list[str] = ["LeagueClient.exe", "LeagueClientUx.exe"]

    logger.info(
        "Will attemt to reconnect in 5 seconds, if the client is still running."
    )
    time.sleep(5)
    if not processes_exists(league_processes):

        # When we close leagueRPC, re-enable the native presence plugin.
        # This prevents users needing to repair the client, the next time they open league.
        if game_path := find_game_path():
            native_presence = check_plugin_status(file_path=game_path, logger=logger)
            if not native_presence and native_presence is not None:
                # If the discord plugin is not present in the manifest.json file, add it.
                if add_plugin(file_path=game_path, plugin_blob=DISCORD_PLUGIN_BLOB):
                    logger.info(
                        "Native League Presence has been re-enabled.", color="yellow"
                    )
                else:
                    logger.error(
                        "Failed to re-enable the Native League Presence.", color="red"
                    )

        # Give people time to read the last messages.
        time.sleep(3)
        await module_data.connector.stop()


@module_data.connector.ws.register(  # type:ignore
    uri="/lol-summoner/v1/current-summoner", event_types=("UPDATE",)
)
async def summoner_updated(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    logger = module_data.logger
    event_data: dict[str, Any] = event.data  # type:ignore

    new_summoner_icon = event_data.get(Summoner.PROFILE_ICON_ID, 0)

    if module_data.client_data.summoner_icon == new_summoner_icon:
        return

    module_data.client_data.summoner_icon = new_summoner_icon
    logger.info("Summoner icon updated.")

    module_data.rpc_updater.delay_update(module_data=module_data, connection=connection)


@module_data.connector.ws.register(  # type:ignore
    uri="/lol-chat/v1/me", event_types=("UPDATE",)
)
async def chat_updated(connection: Connection, event: WebsocketEventResponse) -> None:
    logger = module_data.logger
    event_data: dict[str, Any] = event.data  # type:ignore

    new_status: str = ""

    match event_data[LolChatUser.AVAILABILITY]:
        case LolChatUser.CHAT:
            new_status = LolChatUser.ONLINE.capitalize()
        case LolChatUser.AWAY:
            new_status = LolChatUser.AWAY.capitalize()
        case LolChatUser.DND:
            # We dont care about this.
            return None
        case _:
            return None

    if module_data.client_data.availability == new_status:
        # Chat status is the same, nothing to update
        return None

    module_data.client_data.availability = new_status
    logger.info(f"Status updated to: {new_status}")
    module_data.rpc_updater.delay_update(module_data=module_data, connection=connection)


# Return the selected TFT companion
@module_data.connector.ws.register(  # type:ignore
    uri="/lol-cosmetics/v1/inventories/tft/companions", event_types=("UPDATE",)
)
async def gather_tft_companion_data_updater(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    """
    Gather TFT Companion data from the
    """
    logger = module_data.logger
    event_data: dict[str, Any] = event.data  # type:ignore

    set_tft_companion_data(module_data.client_data, event_data)

    logger.info("TFT Companion updated")
    module_data.rpc_updater.delay_update(module_data=module_data, connection=connection)


@module_data.connector.ws.register(  # type:ignore
    "/lol-gameflow/v1/gameflow-phase", event_types=("UPDATE",)
)
async def gameflow_phase_updated(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    if module_data.client_data.gameflow_phase == event.data:  # type:ignore
        return None

    if GameFlowPhase.GAME_START == event.data:  # type:ignore
        module_data.logger.info("Game is starting. Good luck!", color="blue")
        return None

    module_data.client_data.gameflow_phase = event.data  # type:ignore
    module_data.rpc_updater.delay_update(module_data=module_data, connection=connection)


@module_data.connector.ws.register(  # type:ignore
    uri="/lol-lobby/v2/lobby", event_types=("UPDATE", "CREATE", "DELETE")
)
async def in_lobby(connection: Connection, event: WebsocketEventResponse) -> None:
    event_data: Optional[dict[str, Any]] = getattr(event, "data", None)

    if event_data is None:
        return

    game_config: Optional[dict[str, Any]] = event_data.get("gameConfig")
    if game_config is None:
        return

    # Get queue_id
    module_data.client_data.queue_id = int(game_config["queueId"])
    module_data.client_data.lobby_id = event_data["partyId"]
    module_data.client_data.players = len(event_data["members"])
    module_data.client_data.max_players = int(game_config["maxLobbySize"])
    module_data.client_data.map_id = game_config["mapId"]
    module_data.client_data.gamemode = game_config["gameMode"]
    module_data.client_data.is_custom = game_config["isCustom"]
    if game_config["gameMode"] == "PRACTICETOOL":
        module_data.client_data.is_practice = True
        module_data.client_data.max_players = 1
    else:
        module_data.client_data.is_practice = False

    if module_data.client_data.queue_id == -1:
        # custom game / practice tool / tutorial lobby
        module_data.client_data.queue_detailed_description = ""
        if module_data.client_data.is_practice:
            module_data.client_data.queue_name = "Practice Tool"
        else:
            module_data.client_data.queue_name = "Custom Game"
        module_data.rpc_updater.delay_update(module_data, connection=connection)
        return

    lobby_queue_info_raw: ClientResponse = await connection.request(
        method="GET",
        endpoint="/lol-game-queues/v1/queues/{id}".format_map(
            {"id": module_data.client_data.queue_id}
        ),
    )
    lobby_queue_info: dict[str, Any] = await lobby_queue_info_raw.json()

    module_data.client_data.queue_name = lobby_queue_info["name"]
    module_data.client_data.queue_type = lobby_queue_info["type"]
    module_data.client_data.queue_is_ranked = lobby_queue_info["isRanked"]
    module_data.client_data.queue_detailed_description = lobby_queue_info[
        "detailedDescription"
    ]
    module_data.client_data.queue_description = lobby_queue_info["description"]
    module_data.rpc_updater.delay_update(module_data=module_data, connection=connection)


# ranked stats
@module_data.connector.ws.register(  # type:ignore
    uri="/lol-ranked/v1/current-ranked-stats", event_types=("UPDATE",)
)
async def ranked(connection: Connection, event: WebsocketEventResponse) -> None:
    data: ClientData = module_data.client_data
    event_data: dict[str, Any] = event.data  # type:ignore

    data.summoner_rank = RankedStats.from_map(
        obj_map=event_data,
        ranked_type="RANKED_SOLO_5x5",
    )
    data.summoner_rank_flex = RankedStats.from_map(
        obj_map=event_data,
        ranked_type="RANKED_FLEX_SR",
    )

    data.arena_rank = ArenaStats.from_map(obj_map=event_data)
    data.tft_rank = TFTStats.from_map(obj_map=event_data)

    module_data.rpc_updater.delay_update(module_data, connection)


##### Debug ######
# This will catch all events and print them to the console.


# @module_data.connector.ws.register(  # type:ignore
#     "/",
#     event_types=(
#         "UPDATE",
#         "CREATE",
#         "DELETE",
#     ),
# )
# async def debug(connection: Connection, event: WebsocketEventResponse) -> None:
#     print(f"DEBUG - {event.type}: {event.uri}")


#     import pprint

#     pprint.pprint(event.data)


def start_connector(
    rpc_from_main: Presence,
    cli_args: Namespace,
    logger: RichLogger,
) -> None:
    module_data.rpc = rpc_from_main
    module_data.cli_args = cli_args
    module_data.logger = logger
    module_data.connector.start()
