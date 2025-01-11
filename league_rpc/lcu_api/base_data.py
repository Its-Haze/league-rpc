from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from league_rpc.models.module_data import ModuleData

from aiohttp import ClientResponse
from lcu_driver.connection import Connection

from league_rpc.models.client_data import ArenaStats, ClientData, RankedStats, TFTStats
from league_rpc.models.lcu.current_chat_status import LolChatUser
from league_rpc.models.lcu.current_queue import LolGameQueuesQueue
from league_rpc.models.lcu.current_summoner import Summoner
from league_rpc.models.lcu.gameflow_phase import (
    LolGameflowLobbyStatus,
    LolGameflowPlayerStatus,
)
from league_rpc.utils.const import TFT_COMPANIONS_URL


# Base Data
# Gather base data from the LCU API on startup
async def gather_base_data(connection: Connection, module_data: "ModuleData") -> None:
    data: ClientData = module_data.client_data

    # Epoch time from which league client was started.
    await gather_telemetry_data(connection=connection, data=data)

    await gather_summoner_data(connection=connection, data=data)

    # get Online/Away status
    await gather_chat_status_data(connection=connection, data=data)

    # Get tft companion data
    await gather_tft_companion_data(connection=connection, data=data)

    await gather_ranked_data(connection=connection, data=data)

    await gather_gameflow_data(connection=connection, data=data)

    await gather_lobby_data(connection=connection, data=data)

    if data.queue_id == -1:
        # custom game / practice tool / tutorial lobby
        data.gamemode = "PRACTICETOOL"
        data.queue_detailed_description = ""
        data.map_id = 11
        if data.is_practice:
            data.queue_name = "Practice Tool"
            data.max_players = 1
        else:
            data.queue_name = "Custom Game"

        return

    await gather_queue_data(connection=connection, data=data)


async def gather_queue_data(connection: Connection, data: ClientData) -> None:
    lobby_queue_info_raw: ClientResponse = await connection.request(  # type:ignore
        method="GET", endpoint="/lol-game-queues/v1/queues/" + str(data.queue_id)
    )
    lobby_queue_info: Any = await lobby_queue_info_raw.json()
    data.queue_name = lobby_queue_info[LolGameQueuesQueue.NAME]
    data.queue_type = lobby_queue_info[LolGameQueuesQueue.TYPE]
    data.queue_detailed_description = lobby_queue_info["detailedDescription"]
    data.queue_description = lobby_queue_info["description"]
    data.queue_is_ranked = lobby_queue_info[LolGameQueuesQueue.IS_RANKED]
    data.max_players = int(
        lobby_queue_info[LolGameQueuesQueue.MAXIMUM_PARTICIPANT_LIST_SIZE]
    )
    data.map_id = lobby_queue_info[LolGameQueuesQueue.MAP_ID]
    data.gamemode = lobby_queue_info[LolGameQueuesQueue.GAME_MODE]


async def gather_lobby_data(connection: Connection, data: ClientData) -> None:
    lobby_raw_data: ClientResponse = await connection.request(  # type:ignore
        method="GET", endpoint="/lol-gameflow/v1/gameflow-metadata/player-status"
    )
    lobby_data: dict[str, Any] = await lobby_raw_data.json()

    lobby_status = lobby_data.get(LolGameflowPlayerStatus.CURRENT_LOBBY_STATUS)
    if lobby_status is None:
        return

    data.queue_id = lobby_status[LolGameflowLobbyStatus.QUEUE_ID]
    data.lobby_id = lobby_status[LolGameflowLobbyStatus.LOBBY_ID]
    data.players = len(lobby_status[LolGameflowLobbyStatus.MEMBER_SUMMONER_IDS])
    data.is_practice = lobby_status[LolGameflowLobbyStatus.IS_PRACTICE_TOOL]
    data.is_custom = lobby_status[LolGameflowLobbyStatus.IS_CUSTOM]


async def gather_gameflow_data(connection: Connection, data: ClientData) -> None:
    game_flow_data_raw: ClientResponse = await connection.request(  # type:ignore
        method="GET", endpoint="/lol-gameflow/v1/gameflow-phase"
    )
    game_flow_data: str = await game_flow_data_raw.json()
    data.gameflow_phase = game_flow_data


async def gather_ranked_data(connection: Connection, data: ClientData) -> None:
    ranked_data_raw: ClientResponse = await connection.request(  # type:ignore
        method="GET", endpoint="/lol-ranked/v1/current-ranked-stats/"
    )
    ranked_data: dict[str, Any] = await ranked_data_raw.json()

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


async def gather_tft_companion_data(connection: Connection, data: ClientData) -> None:
    """
    Gather TFT Companion data from the LCU API.
    """
    tft_companion_data_raw: ClientResponse = await connection.request(  # type:ignore
        method="GET", endpoint="/lol-cosmetics/v1/inventories/tft/companions"
    )

    tft_companion_data_json = await tft_companion_data_raw.json()

    set_tft_companion_data(data, tft_companion_data_json)


def set_tft_companion_data(
    data: ClientData,
    companion_data_json: dict[str, Any],
) -> None:
    """Set TFT Companion data from the LCU API."""
    companion_data = companion_data_json["selectedLoadoutItem"]

    data.tft_companion_id = companion_data["itemId"]

    # Get the full path of the TFT Companion icon
    full_companion_icon_path: str = companion_data["loadoutsIcon"]

    # Only get the final part and add it to community dragon URL
    final_companion_icon_path: str = full_companion_icon_path.split("ASSETS/")[
        -1
    ].lower()

    # Set the TFT Companion icon URL
    data.tft_companion_icon = f"{TFT_COMPANIONS_URL}/{final_companion_icon_path}"
    data.tft_companion_name = companion_data["name"]
    data.tft_companion_description = companion_data["description"]


async def gather_chat_status_data(connection: Connection, data: ClientData) -> None:
    chat_data_raw: ClientResponse = await connection.request(  # type:ignore
        method="GET", endpoint="/lol-chat/v1/me"
    )
    chat_data: dict[str, Any] = await chat_data_raw.json()

    match chat_data[LolChatUser.AVAILABILITY]:
        case LolChatUser.CHAT:
            data.availability = LolChatUser.ONLINE.capitalize()
        case LolChatUser.AWAY:
            data.availability = LolChatUser.AWAY.capitalize()
        case _:
            ...


async def gather_summoner_data(connection: Connection, data: ClientData) -> None:
    summoner_data_raw: ClientResponse = await connection.request(  # type:ignore
        method="GET", endpoint="/lol-summoner/v1/current-summoner"
    )
    summoner_data = await summoner_data_raw.json()

    data.summoner_icon = summoner_data[Summoner.PROFILE_ICON_ID]


async def gather_telemetry_data(connection: Connection, data: ClientData) -> None:
    application_start_time_raw: ClientResponse = (
        await connection.request(  # type:ignore
            method="GET", endpoint="/telemetry/v1/application-start-time"
        )
    )
    application_start_time: int = await application_start_time_raw.json()
    data.application_start_time = application_start_time
