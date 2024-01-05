from argparse import Namespace
from typing import Any, Optional

from lcu_driver.connection import Connection
from lcu_driver.events.responses import WebsocketEventResponse
from pypresence import Presence

from league_rpc_linux.colors import Colors
from league_rpc_linux.lcu_api.base_data import gather_base_data
from league_rpc_linux.models.client_data import RankedStats
from league_rpc_linux.models.lcu.current_chat_status import LolChatUser
from league_rpc_linux.models.lcu.current_lobby import (
    LolLobbyLobbyDto,
    LolLobbyLobbyGameConfigDto,
)
from league_rpc_linux.models.lcu.current_queue import LolGameQueuesQueue
from league_rpc_linux.models.lcu.current_ranked_stats import (
    LolRankedRankedQueueStats,
    LolRankedRankedStats,
)
from league_rpc_linux.models.lcu.current_summoner import Summoner
from league_rpc_linux.models.module_data import ModuleData
from league_rpc_linux.models.rpc_updater import RPCUpdater

module_data = ModuleData()
rpc_updater = RPCUpdater()

## WS Events ##


@module_data.connector.ready  # type:ignore
async def connect(connection: Connection):
    print(f"{Colors.green}LCU API is ready.{Colors.reset}")

    await gather_base_data(connection, module_data)
    rpc_updater.delay_update(module_data)


@module_data.connector.close  # type:ignore
async def disconnect(_: Connection):
    print("LCU API is closed.")


@module_data.connector.ws.register(  # type:ignore
    "/lol-summoner/v1/current-summoner", event_types=("UPDATE",)
)
async def summoner_updated(_: Connection, event: WebsocketEventResponse) -> None:
    print("Summoner has been updated.")
    data = module_data.client_data
    event_data: dict[str, Any] = event.data  # type:ignore

    data.summoner_name = event_data[Summoner.DISPLAY_NAME]
    data.summoner_level = event_data[Summoner.SUMMONER_LEVEL]
    data.summoner_id = event_data[Summoner.SUMMONER_ID]
    data.summoner_icon = event_data[Summoner.PROFILE_ICON_ID]

    rpc_updater.delay_update(module_data)


@module_data.connector.ws.register(  # type:ignore
    "/lol-chat/v1/me", event_types=("UPDATE",)
)
async def chat_updated(_: Connection, event: WebsocketEventResponse) -> None:
    data = module_data.client_data
    event_data: dict[str, Any] = event.data  # type:ignore

    print("Chat has been updated.")

    match event_data[LolChatUser.AVAILABILITY]:
        case LolChatUser.CHAT:
            data.availability = LolChatUser.ONLINE.capitalize()

        case LolChatUser.AWAY:
            data.availability = LolChatUser.AWAY.capitalize()
        case _:
            ...
    rpc_updater.delay_update(module_data)


@module_data.connector.ws.register(  # type:ignore
    "/lol-gameflow/v1/gameflow-phase", event_types=("UPDATE",)
)
async def gameflow_phase_updated(_: Connection, event: WebsocketEventResponse) -> None:
    data = module_data.client_data
    event_data: Any = event.data  # type:ignore
    print("Gameflow Phase has been updated.")

    data.gameflow_phase = event_data  # returns plain string of the phase
    rpc_updater.delay_update(module_data)


# could be used for lobby instead: /lol-gameflow/v1/gameflow-metadata/player-status
@module_data.connector.ws.register(  # type:ignore
    "/lol-lobby/v2/lobby", event_types=("UPDATE", "CREATE", "DELETE")
)
async def in_lobby(connection: Connection, event: WebsocketEventResponse) -> None:
    data = module_data.client_data
    event_data: Optional[dict[str, Any]] = event.data  # type:ignore
    print(f"Lobby Data - {event.type}")  # type:ignore

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

    lobby_queue_info_raw = await connection.request(
        "GET", "/lol-game-queues/v1/queues/{id}".format_map({"id": data.queue_id})
    )
    lobby_queue_info = await lobby_queue_info_raw.json()

    data.queue = lobby_queue_info[LolGameQueuesQueue.NAME]
    data.queue_is_ranked = lobby_queue_info[LolGameQueuesQueue.IS_RANKED]

    rpc_updater.delay_update(module_data)


# ranked stats
@module_data.connector.ws.register(  # type:ignore
    "/lol-ranked/v1/current-ranked-stats/", event_types=("UPDATE",)
)
async def ranked(_: Connection, event: WebsocketEventResponse) -> None:
    data = module_data.client_data

    event_data: dict[str, Any] = event.data  # type:ignore
    solo_queue = "RANKED_SOLO_5x5"
    flex_queue = "RANKED_FLEX_SR"

    data.summoner_rank = RankedStats(
        division=event_data[LolRankedRankedStats.QUEUE_MAP][solo_queue][
            LolRankedRankedQueueStats.DIVISION
        ],
        tier=event_data[LolRankedRankedStats.QUEUE_MAP][solo_queue][
            LolRankedRankedQueueStats.TIER
        ],
        league_points=event_data[LolRankedRankedStats.QUEUE_MAP][solo_queue][
            LolRankedRankedQueueStats.LEAGUE_POINTS
        ],
    )
    data.summoner_rank_flex = RankedStats(
        division=event_data[LolRankedRankedStats.QUEUE_MAP][flex_queue][
            LolRankedRankedQueueStats.DIVISION
        ],
        tier=event_data[LolRankedRankedStats.QUEUE_MAP][flex_queue][
            LolRankedRankedQueueStats.TIER
        ],
        league_points=event_data[LolRankedRankedStats.QUEUE_MAP][flex_queue][
            LolRankedRankedQueueStats.LEAGUE_POINTS
        ],
    )

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
    print("Starting LCU API.")
    module_data.connector.start()
