from lcu_driver.connection import Connection

from league_rpc_linux.lcu_api.current_chat_status import LolChatUser
from league_rpc_linux.lcu_api.current_queue import LolGameQueuesQueue
from league_rpc_linux.lcu_api.current_ranked_stats import (
    LolRankedRankedQueueStats,
    LolRankedRankedStats,
)
from league_rpc_linux.lcu_api.current_summoner import Summoner
from league_rpc_linux.lcu_api.gameflow_phase import (
    GameFlowPhase,
    LolGameflowLobbyStatus,
    LolGameflowPlayerStatus,
)
from league_rpc_linux.models.module_data import ModuleData


# Base Data
# Gather base data from the LCU API on startup
async def gather_base_data(connection: Connection, module_data: ModuleData):
    print("Gathering base data.")
    data = module_data.client_data
    summoner_data_raw = await connection.request(
        "GET", "/lol-summoner/v1/current-summoner"
    )
    summoner_data = await summoner_data_raw.json()
    data.summoner_name = summoner_data[Summoner.DISPLAY_NAME]
    data.summoner_level = summoner_data[Summoner.SUMMONER_LEVEL]
    data.summoner_id = summoner_data[Summoner.SUMMONER_ID]
    data.summoner_icon = summoner_data[Summoner.PROFILE_ICON_ID]

    # get Online/Away status
    chat_data_raw = await connection.request("GET", "/lol-chat/v1/me")
    chat_data = await chat_data_raw.json()
    match chat_data[LolChatUser.AVAILABILITY]:
        case LolChatUser.CHAT:
            data.availability = LolChatUser.ONLINE.capitalize()
        case LolChatUser.AWAY:
            data.availability = LolChatUser.AWAY.capitalize()
        case _:
            ...

    ranked_data_raw = await connection.request(
        "GET", "/lol-ranked/v1/current-ranked-stats/"
    )
    ranked_data = await ranked_data_raw.json()

    data.summoner_rank = (
        f"{ranked_data[LolRankedRankedStats.HIGHEST_RANKED_ENTRY][LolRankedRankedQueueStats.TIER]} "
        f"{ranked_data[LolRankedRankedStats.HIGHEST_RANKED_ENTRY][LolRankedRankedQueueStats.DIVISION]}: "
        f"{str(ranked_data[LolRankedRankedStats.HIGHEST_RANKED_ENTRY][LolRankedRankedQueueStats.LEAGUE_POINTS])} LP"
    )
    game_flow_data_raw = await connection.request(
        "GET", "/lol-gameflow/v1/gameflow-phase"
    )
    game_flow_data = await game_flow_data_raw.json()
    data.gameflow_phase = game_flow_data

    if data.gameflow_phase == GameFlowPhase.IN_PROGRESS:
        # In Game
        return

    if data.gameflow_phase == GameFlowPhase.NONE:
        # In Client
        return

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
