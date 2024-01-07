from lcu_driver.connection import Connection

from league_rpc_linux.models.client_data import (
    ArenaStats,
    ClientData,
    RankedStats,
    TFTStats,
)
from league_rpc_linux.models.lcu.current_chat_status import LolChatUser
from league_rpc_linux.models.lcu.current_queue import LolGameQueuesQueue
from league_rpc_linux.models.lcu.current_summoner import Summoner
from league_rpc_linux.models.lcu.gameflow_phase import (
    GameFlowPhase,
    LolGameflowLobbyStatus,
    LolGameflowPlayerStatus,
)
from league_rpc_linux.models.module_data import ModuleData


# Base Data
# Gather base data from the LCU API on startup
async def gather_base_data(connection: Connection, module_data: ModuleData):
    data = module_data.client_data

    # Epoch time from which league client was started.
    await gather_telemetry_data(connection, data)

    await gather_summoner_data(connection, data)

    # get Online/Away status
    await gather_chat_status_data(connection, data)

    await gather_ranked_data(connection, data)

    await gather_gameflow_data(connection, data)

    if data.gameflow_phase == GameFlowPhase.IN_PROGRESS:
        # In Game
        return

    if data.gameflow_phase == GameFlowPhase.NONE:
        # In Client
        return

    await gather_lobby_data(connection, data)

    if data.queue_id == -1:
        # custom game / practice tool / tutorial lobby
        data.gamemode = "PRACTICETOOL"
        data.map_id = 11
        if data.is_practice:
            data.queue = "Practice Tool"
            data.max_players = 1
        else:
            data.queue = "Custom Game"

        return

    await gather_queue_data(connection, data)


async def gather_queue_data(connection: Connection, data: ClientData):
    lobby_queue_info_raw = await connection.request(
        "GET", "/lol-game-queues/v1/queues/" + str(data.queue_id)
    )
    lobby_queue_info = await lobby_queue_info_raw.json()
    data.queue = lobby_queue_info[LolGameQueuesQueue.NAME]
    data.max_players = int(
        lobby_queue_info[LolGameQueuesQueue.MAXIMUM_PARTICIPANT_LIST_SIZE]
    )
    data.map_id = lobby_queue_info[LolGameQueuesQueue.MAP_ID]
    data.gamemode = lobby_queue_info[LolGameQueuesQueue.GAME_MODE]
    data.queue_is_ranked = lobby_queue_info[LolGameQueuesQueue.IS_RANKED]


async def gather_lobby_data(connection: Connection, data: ClientData):
    lobby_raw_data = await connection.request(
        "GET", "/lol-gameflow/v1/gameflow-metadata/player-status"
    )
    lobby_data = await lobby_raw_data.json()

    data.queue_id = lobby_data[LolGameflowPlayerStatus.CURRENT_LOBBY_STATUS][
        LolGameflowLobbyStatus.QUEUE_ID
    ]
    data.lobby_id = lobby_data[LolGameflowPlayerStatus.CURRENT_LOBBY_STATUS][
        LolGameflowLobbyStatus.LOBBY_ID
    ]
    data.players = len(
        lobby_data[LolGameflowPlayerStatus.CURRENT_LOBBY_STATUS][
            LolGameflowLobbyStatus.MEMBER_SUMMONER_IDS
        ]
    )
    data.is_practice = lobby_data[LolGameflowPlayerStatus.CURRENT_LOBBY_STATUS][
        LolGameflowLobbyStatus.IS_PRACTICE_TOOL
    ]
    data.is_custom = lobby_data[LolGameflowPlayerStatus.CURRENT_LOBBY_STATUS][
        LolGameflowLobbyStatus.IS_CUSTOM
    ]


async def gather_gameflow_data(connection: Connection, data: ClientData):
    game_flow_data_raw = await connection.request(
        "GET", "/lol-gameflow/v1/gameflow-phase"
    )
    game_flow_data = await game_flow_data_raw.json()
    data.gameflow_phase = game_flow_data


async def gather_ranked_data(connection: Connection, data: ClientData):
    ranked_data_raw = await connection.request(
        "GET", "/lol-ranked/v1/current-ranked-stats/"
    )
    ranked_data = await ranked_data_raw.json()

    data.summoner_rank = RankedStats.from_map(
        obj_map=ranked_data,
        ranked_type="RANKED_SOLO_5x5",
    )
    data.summoner_rank_flex = RankedStats.from_map(
        obj_map=ranked_data,
        ranked_type="RANKED_FLEX_SR",
    )

    data.arena_rank = ArenaStats.from_map(obj_map=ranked_data)
    data.tft_rank = TFTStats.from_map(obj_map=ranked_data)


async def gather_chat_status_data(connection: Connection, data: ClientData):
    chat_data_raw = await connection.request("GET", "/lol-chat/v1/me")
    chat_data = await chat_data_raw.json()

    match chat_data[LolChatUser.AVAILABILITY]:
        case LolChatUser.CHAT:
            data.availability = LolChatUser.ONLINE.capitalize()
        case LolChatUser.AWAY:
            data.availability = LolChatUser.AWAY.capitalize()
        case _:
            ...


async def gather_summoner_data(connection: Connection, data: ClientData):
    summoner_data_raw = await connection.request(
        "GET", "/lol-summoner/v1/current-summoner"
    )
    summoner_data = await summoner_data_raw.json()

    data.summoner_name = summoner_data[Summoner.DISPLAY_NAME]
    data.summoner_level = summoner_data[Summoner.SUMMONER_LEVEL]
    data.summoner_id = summoner_data[Summoner.SUMMONER_ID]
    data.summoner_icon = summoner_data[Summoner.PROFILE_ICON_ID]


async def gather_telemetry_data(connection: Connection, data: ClientData):
    application_start_time_raw = await connection.request(
        "GET", "/telemetry/v1/application-start-time"
    )
    application_start_time: int = await application_start_time_raw.json()
    data.application_start_time = application_start_time
