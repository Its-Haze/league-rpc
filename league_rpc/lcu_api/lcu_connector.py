import time
from argparse import Namespace
from typing import Any, Optional

from aiohttp import ClientResponse
from lcu_driver.connection import Connection  # type:ignore

from lcu_driver.events.responses import WebsocketEventResponse  # type:ignore
from pypresence import Presence  # type:ignore


from league_rpc.disable_native_rpc.disable import check_plugin_status, find_game_path
from league_rpc.lcu_api.base_data import gather_base_data
from league_rpc.models.client_data import ArenaStats, ClientData, RankedStats, TFTStats
from league_rpc.models.lcu.current_chat_status import LolChatUser

from league_rpc.models.lcu.current_summoner import Summoner
from league_rpc.models.module_data import ModuleData
from league_rpc.models.rpc_updater import RPCUpdater
from league_rpc.utils.color import Color
from league_rpc.utils.const import (
    TFT_COMPANIONS_URL,
)
from pprint import pprint
import rich

module_data = ModuleData()
rpc_updater = RPCUpdater()

## WS Events ##


@module_data.connector.ready  # type:ignore
async def connect(connection: Connection) -> None:
    print(f"{Color.green}Successfully connected to the League Client API.{Color.reset}")
    time.sleep(2)  # Give the client some time to load

    print(f"\n{Color.orange}Gathering base data.{Color.reset}")
    time.sleep(2)
    await gather_base_data(connection=connection, module_data=module_data)

    print(f"{Color.green}Successfully gathered base data.{Color.reset}")

    print(f"\n{Color.orange}Updating Discord rpc with base data{Color.reset}")
    rpc_updater.delay_update(module_data=module_data, connection=connection)
    print(f"{Color.green}Discord RPC successfully updated{Color.reset}")

    print(f"\n{Color.cyan}LeagueRPC is ready{Color.reset}")

    if game_path := find_game_path():
        check_plugin_status(file_path=game_path)


@module_data.connector.close  # type:ignore
async def disconnect(_: Connection) -> None:
    print(f"{Color.red}Disconnected from the League Client API.{Color.reset}")


@module_data.connector.ws.register(  # type:ignore
    uri="/lol-summoner/v1/current-summoner", event_types=("UPDATE",)
)
async def summoner_updated(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    print("Summoner updated")
    data: ClientData = module_data.client_data
    event_data: dict[str, Any] = event.data  # type:ignore

    new_summoner_icon = event_data.get(Summoner.PROFILE_ICON_ID, 0)

    if data.summoner_icon == new_summoner_icon:
        print("Summoner icon is the same, nothing to update")
        return

    data.summoner_icon = new_summoner_icon

    print("Updating RPC from summoner_updated")
    rpc_updater.delay_update(module_data=module_data, connection=connection)


@module_data.connector.ws.register(  # type:ignore
    uri="/lol-chat/v1/me", event_types=("UPDATE",)
)
async def chat_updated(connection: Connection, event: WebsocketEventResponse) -> None:
    print("Online status updated")
    data: ClientData = module_data.client_data
    event_data: dict[str, Any] = event.data  # type:ignore

    match event_data[LolChatUser.AVAILABILITY]:
        case LolChatUser.CHAT:
            data.availability = LolChatUser.ONLINE.capitalize()

        case LolChatUser.AWAY:
            data.availability = LolChatUser.AWAY.capitalize()
        case _:
            print("Unknown chat status")
            return None

    print("Updating RPC from chat_updated")
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
    print("TFT Companion data updated")
    data: ClientData = module_data.client_data
    event_data: dict[str, Any] = event.data  # type:ignore

    companion_data = event_data.get("selectedLoadoutItem")
    if not companion_data:
        print("No TFT Companion data found.")
        return

    data.tft_companion_id = companion_data["itemId"]

    companion_icon_path: str = companion_data["loadoutsIcon"]
    companion_file_name: str = companion_icon_path.split("/")[-1].lower()

    # Set the TFT Companion icon URL
    data.tft_companion_icon = f"{TFT_COMPANIONS_URL}/{companion_file_name}"

    data.tft_companion_name = companion_data["name"]
    data.tft_companion_description = companion_data["description"]

    print("Updating RPC from gather_tft_companion_data_updater")
    rpc_updater.delay_update(module_data=module_data, connection=connection)


@module_data.connector.ws.register(  # type:ignore
    "/lol-gameflow/v1/gameflow-phase", event_types=("UPDATE",)
)
async def gameflow_phase_updated(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    print("Gameflow phase updated")

    if module_data.client_data.gameflow_phase == event.data:  # type:ignore
        print("Gameflow phase is the same, nothing to update")
        return None

    module_data.client_data.gameflow_phase = event.data  # type:ignore

    print("Updating RPC from gameflow_phase_updated")
    rpc_updater.delay_update(module_data=module_data, connection=connection)


# could be used for lobby instead: /lol-gameflow/v1/gameflow-metadata/player-status
@module_data.connector.ws.register(  # type:ignore
    uri="/lol-lobby/v2/lobby", event_types=("UPDATE", "CREATE", "DELETE")
)
async def in_lobby(connection: Connection, event: WebsocketEventResponse) -> None:

    print("Lobby updated")
    event_data: Optional[dict[str, Any]] = getattr(event, "data", None)

    if event_data is None:
        # Make an early return if data is not present in the event.
        # Probably us leaving the lobby.
        print("Updating RPC from in_lobby - event_data is None")
        # rpc_updater.delay_update(module_data, connection=connection)
        return

    game_config: Optional[dict[str, Any]] = event_data.get("gameConfig")
    if game_config is None:
        print("Updating RPC from in_lobby - game_config is None")
        # rpc_updater.delay_update(module_data, connection=connection)
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
        print("Updating RPC from in_lobby - custom game")
        rpc_updater.delay_update(module_data, connection=connection)
        return

    lobby_queue_info_raw: ClientResponse = await connection.request(
        method="GET",
        endpoint="/lol-game-queues/v1/queues/{id}".format_map(
            {"id": module_data.client_data.queue_id}
        ),
    )
    lobby_queue_info: dict[str, Any] = await lobby_queue_info_raw.json()

    # pprint("Lobby queue info")
    # pprint(lobby_queue_info)
    module_data.client_data.queue_name = lobby_queue_info["name"]
    module_data.client_data.queue_type = lobby_queue_info["type"]
    module_data.client_data.queue_is_ranked = lobby_queue_info["isRanked"]
    module_data.client_data.queue_detailed_description = lobby_queue_info[
        "detailedDescription"
    ]
    module_data.client_data.queue_description = lobby_queue_info["description"]
    print("Updating RPC from in_lobby - at the bottom")
    rpc_updater.delay_update(module_data=module_data, connection=connection)


# ranked stats
@module_data.connector.ws.register(  # type:ignore
    uri="/lol-ranked/v1/current-ranked-stats", event_types=("UPDATE",)
)
async def ranked(connection: Connection, event: WebsocketEventResponse) -> None:
    print("Ranked stats updated")
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

    print("Updating RPC from ranked - at the bottom")
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


def start_connector(rpc_from_main: Presence, cli_args: Namespace) -> None:
    module_data.rpc = rpc_from_main
    module_data.cli_args = cli_args
    module_data.connector.start()
