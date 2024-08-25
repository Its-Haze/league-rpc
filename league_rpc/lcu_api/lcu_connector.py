import time
from argparse import Namespace
from typing import Any, Optional

from aiohttp import ClientResponse
from lcu_driver.connection import Connection  # type:ignore

from lcu_driver.events.responses import WebsocketEventResponse  # type:ignore
from pypresence import Presence  # type:ignore


from league_rpc.disable_native_rpc.disable import check_plugin_status, find_game_path
from league_rpc.lcu_api.base_data import gather_base_data, set_tft_companion_data
from league_rpc.logger.richlogger import RichLogger
from league_rpc.models.client_data import ArenaStats, ClientData, RankedStats, TFTStats
from league_rpc.models.lcu.current_chat_status import LolChatUser

from league_rpc.models.lcu.current_summoner import Summoner
from league_rpc.models.module_data import ModuleData
from league_rpc.models.rpc_updater import RPCUpdater
from league_rpc.utils.const import (
    TFT_COMPANIONS_URL,
)

module_data = ModuleData()
rpc_updater = RPCUpdater()


@module_data.connector.ready  # type:ignore
async def connect(connection: Connection) -> None:
    logger = module_data.logger
    logger.info("Connected to the League Client API.")
    time.sleep(2)  # Give the client some time to load

    logger.start_progress_bar(name="Gathering Base Data")

    time.sleep(2)
    await gather_base_data(connection=connection, module_data=module_data)

    logger.info("Successfully gathered base data.")

    logger.update_progress_bar(advance=50)

    logger.info("Updating Discord rpc with base data", color="yellow")
    rpc_updater.delay_update(module_data=module_data, connection=connection)
    logger.info("Discord RPC successfully updated")
    logger.update_progress_bar(advance=50)

    logger.stop_progress_bar()

    logger.info("LeagueRPC is ready!", color="cyan")
    if game_path := find_game_path():
        check_plugin_status(file_path=game_path, logger=logger)


@module_data.connector.close  # type:ignore
async def disconnect(_: Connection) -> None:
    logger = module_data.logger
    logger.error("Disconnected from the League Client API.", color="red")


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

    rpc_updater.delay_update(module_data=module_data, connection=connection)


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
    logger.info("Online/Away status updated")
    rpc_updater.delay_update(module_data=module_data, connection=connection)


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

    logger.info("TFT Companion data updated.")
    rpc_updater.delay_update(module_data=module_data, connection=connection)


@module_data.connector.ws.register(  # type:ignore
    "/lol-gameflow/v1/gameflow-phase", event_types=("UPDATE",)
)
async def gameflow_phase_updated(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    if module_data.client_data.gameflow_phase == event.data:  # type:ignore
        return None

    module_data.client_data.gameflow_phase = event.data  # type:ignore

    rpc_updater.delay_update(module_data=module_data, connection=connection)


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
        rpc_updater.delay_update(module_data, connection=connection)
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
    rpc_updater.delay_update(module_data=module_data, connection=connection)


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

    rpc_updater.delay_update(module_data, connection)


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
