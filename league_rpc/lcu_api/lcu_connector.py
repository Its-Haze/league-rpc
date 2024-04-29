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
from league_rpc.models.lcu.current_lobby import (
    LolLobbyLobbyDto,
    LolLobbyLobbyGameConfigDto,
)
from league_rpc.models.lcu.current_queue import LolGameQueuesQueue
from league_rpc.models.lcu.current_summoner import Summoner
from league_rpc.models.module_data import ModuleData
from league_rpc.models.rpc_updater import RPCUpdater
from league_rpc.utils.color import Color

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
    rpc_updater.delay_update(module_data=module_data)
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
async def summoner_updated(_: Connection, event: WebsocketEventResponse) -> None:
    data: ClientData = module_data.client_data
    event_data: dict[str, Any] = event.data  # type:ignore

    data.summoner_name = event_data[Summoner.DISPLAY_NAME]
    data.summoner_level = event_data[Summoner.SUMMONER_LEVEL]
    data.summoner_id = event_data[Summoner.SUMMONER_ID]
    data.summoner_icon = event_data[Summoner.PROFILE_ICON_ID]

    rpc_updater.delay_update(module_data=module_data)


@module_data.connector.ws.register(  # type:ignore
    uri="/lol-chat/v1/me", event_types=("UPDATE",)
)
async def chat_updated(_: Connection, event: WebsocketEventResponse) -> None:
    data: ClientData = module_data.client_data
    event_data: dict[str, Any] = event.data  # type:ignore

    match event_data[LolChatUser.AVAILABILITY]:
        case LolChatUser.CHAT:
            data.availability = LolChatUser.ONLINE.capitalize()

        case LolChatUser.AWAY:
            data.availability = LolChatUser.AWAY.capitalize()
        case _:
            ...
    rpc_updater.delay_update(module_data=module_data)


@module_data.connector.ws.register(  # type:ignore
    "/lol-gameflow/v1/gameflow-phase", event_types=("UPDATE",)
)
async def gameflow_phase_updated(_: Connection, event: WebsocketEventResponse) -> None:
    data: ClientData = module_data.client_data
    event_data: Any = event.data  # type:ignore

    data.gameflow_phase = event_data  # returns plain string of the phase
    rpc_updater.delay_update(module_data=module_data)


# could be used for lobby instead: /lol-gameflow/v1/gameflow-metadata/player-status
@module_data.connector.ws.register(  # type:ignore
    uri="/lol-lobby/v2/lobby", event_types=("UPDATE", "CREATE", "DELETE")
)
async def in_lobby(connection: Connection, event: WebsocketEventResponse) -> None:
    data: ClientData = module_data.client_data
    event_data: Optional[dict[str, Any]] = event.data  # type:ignore

    if event_data is None:
        # Make an early return if data is not present in the event.
        return

    data.queue_id = int(
        event_data[LolLobbyLobbyDto.GAME_CONFIG][
            LolLobbyLobbyGameConfigDto.QUEUE_ID
        ]  # type:ignore
    )
    data.lobby_id = event_data[LolLobbyLobbyDto.PARTY_ID]
    data.players = len(event_data[LolLobbyLobbyDto.MEMBERS])  # type:ignore
    data.max_players = int(
        event_data[LolLobbyLobbyDto.GAME_CONFIG][  # type:ignore
            LolLobbyLobbyGameConfigDto.MAX_LOBBY_SIZE
        ]
    )
    data.map_id = event_data[LolLobbyLobbyDto.GAME_CONFIG][
        LolLobbyLobbyGameConfigDto.MAP_ID
    ]
    data.gamemode = event_data[LolLobbyLobbyDto.GAME_CONFIG][
        LolLobbyLobbyGameConfigDto.GAME_MODE
    ]
    data.is_custom = event_data[LolLobbyLobbyDto.GAME_CONFIG][
        LolLobbyLobbyGameConfigDto.IS_CUSTOM
    ]
    if (
        event_data[LolLobbyLobbyDto.GAME_CONFIG][LolLobbyLobbyGameConfigDto.GAME_MODE]
        == "PRACTICETOOL"
    ):
        data.is_practice = True
        data.max_players = 1
    else:
        data.is_practice = False

    if data.queue_id == -1:
        # custom game / practice tool / tutorial lobby
        if data.is_practice:
            data.queue = "Practice Tool"
        else:
            data.queue = "Custom Game"
        rpc_updater.delay_update(module_data)
        return

    lobby_queue_info_raw: ClientResponse = await connection.request(
        method="GET",
        endpoint="/lol-game-queues/v1/queues/{id}".format_map({"id": data.queue_id}),
    )
    lobby_queue_info: dict[str, Any] = await lobby_queue_info_raw.json()

    data.queue = lobby_queue_info[LolGameQueuesQueue.NAME]
    data.queue_type = lobby_queue_info[LolGameQueuesQueue.TYPE]
    data.queue_is_ranked = lobby_queue_info[LolGameQueuesQueue.IS_RANKED]

    rpc_updater.delay_update(module_data=module_data)


# ranked stats
@module_data.connector.ws.register(  # type:ignore
    uri="/lol-ranked/v1/current-ranked-stats", event_types=("UPDATE",)
)
async def ranked(_: Connection, event: WebsocketEventResponse) -> None:
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

    rpc_updater.delay_update(module_data)


###### Debug ######
# This will catch all events and print them to the console.

# @module_data.connector.ws.register(  # type:ignore
#    "/", event_types=("UPDATE", "CREATE", "DELETE")
# )
# async def debug(connection: Connection, event: WebsocketEventResponse) -> None:
#    print(f"DEBUG - {event.type}: {event.uri}")


def start_connector(rpc_from_main: Presence, cli_args: Namespace) -> None:
    module_data.rpc = rpc_from_main
    module_data.cli_args = cli_args
    module_data.connector.start()
