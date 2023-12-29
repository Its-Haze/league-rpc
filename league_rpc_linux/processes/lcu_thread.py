import time
from threading import Timer
from dataclasses import dataclass, field

import lcu_driver
from lcu_driver.connection import Connection
from lcu_driver.events.responses import WebsocketEventResponse
from pypresence import Presence

from league_rpc_linux.colors import Colors
from league_rpc_linux.const import (
    BASE_MAP_ICON_URL,
    GAME_MODE_CONVERT_MAP,
    LEAGUE_OF_LEGENDS_LOGO,
    MAP_ICON_CONVERT_MAP,
    PROFILE_ICON_BASE_URL,
    SMALL_TEXT,
)


@dataclass
class ClientData:
    availability: str = "Online"  # "Online", "Away
    gamemode: str = None
    gameflow_phase: str = "None"  # None, Lobby, Matchmaking, ReadyCheck, ChampSelect, InProgress, WaitingForStats, EndOfGame
    is_custom: bool = False
    is_practice: bool = False
    lobby_id: str = None  # unique lobby id
    map_id: int = None  # 11, 12, 21, 22, 30
    max_players: int = None  # max players in lobby
    players: int = None  # players in lobby
    queue: str = None
    queue_id: int = -1
    queue_is_ranked: bool = False
    summoner_icon: int = None
    summoner_id: str = None
    summoner_level: int = None
    summoner_name: str = None
    summoner_rank: str = None


## Timers ##


# As some events are called multiple times, we should limit the amount of updates to the RPC.
# Collect update events for 1 second and then update the RPC.
class RPCUpdater:
    def __init__(self):
        self.scheduled_update = False

    def delay_update(self):
        if not self.scheduled_update:
            self.scheduled_update = True
            Timer(1.0, self.update_rpc_and_reset_flag).start()

    def update_rpc_and_reset_flag(self):
        module_data.update_rpc()  # Assuming update_rpc is defined elsewhere
        self.scheduled_update = False


# contains module internal data
class ModuleData:
    connector: lcu_driver.Connector = lcu_driver.Connector()
    rpc: Presence = None
    rpc_updater: RPCUpdater = RPCUpdater()
    client_data: ClientData = ClientData()
    update_rpc: callable = None


module_data = ModuleData()

## WS Events ##


@module_data.connector.ready
async def connect(connection: Connection):
    print(f"{Colors.green}LCU API is ready.{Colors.reset}")

    await gather_base_data(connection)
    module_data.rpc_updater.delay_update()


@module_data.connector.close
async def disconnect(connection: Connection):
    print("LCU API is closed.")


@module_data.connector.ws.register(
    "/lol-summoner/v1/current-summoner", event_types=("UPDATE",)
)
async def summoner_updated(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    print("Summoner has been updated.")
    data = module_data.client_data
    data.summoner_name = event.data["displayName"]
    data.summoner_level = event.data["summonerLevel"]
    data.summoner_id = event.data["summonerId"]
    data.summoner_icon = event.data["profileIconId"]

    module_data.rpc_updater.delay_update()


@module_data.connector.ws.register("/lol-chat/v1/me", event_types=("UPDATE",))
async def chat_updated(connection: Connection, event: WebsocketEventResponse) -> None:
    data = module_data.client_data
    print("Chat has been updated.")
    if event.data["availability"] == "chat":
        data.availability = "Online"
    if event.data["availability"] == "away":
        data.availability = "Away"

    module_data.rpc_updater.delay_update()


@module_data.connector.ws.register(
    "/lol-gameflow/v1/gameflow-phase", event_types=("UPDATE",)
)
async def gameflow_phase_updated(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    data = module_data.client_data
    print("Gameflow Phase has been updated.")
    data.gameflow_phase = event.data  # returns plain string of the phase

    module_data.rpc_updater.delay_update()


# could be used for lobby instead: /lol-gameflow/v1/gameflow-metadata/player-status
@module_data.connector.ws.register(
    "/lol-lobby/v2/lobby", event_types=("UPDATE", "CREATE", "DELETE")
)
async def in_lobby(connection: Connection, event: WebsocketEventResponse) -> None:
    data = module_data.client_data
    print(f"Lobby Data - {event.type}")

    if event.type == "Delete":
        module_data.rpc_updater.delay_update()
        return

    if event.type == "Create":
        # wait for Update Call, as create does not contain data
        return

    data.queue_id = event.data["gameConfig"]["queueId"]
    data.lobby_id = event.data["partyId"]
    data.players = len(event.data["members"])
    data.max_players = int(event.data["gameConfig"]["maxLobbySize"])
    data.map_id = event.data["gameConfig"]["mapId"]
    data.gamemode = event.data["gameConfig"]["gameMode"]
    data.is_custom = event.data["gameConfig"]["isCustom"]
    if event.data["gameConfig"]["gameMode"] == "PRACTICETOOL":
        data.is_practice = True
        data.max_players = 1

    if data.queue_id == -1:
        # custom game / practice tool / tutorial lobby
        if data.is_practice:
            data.queue = "Practice Tool"
        else:
            data.queue = "Custom Game"
        module_data.rpc_updater.delay_update()
        return

    lobbyQueueInfoRaw = await connection.request(
        "GET", "/lol-game-queues/v1/queues/" + str(data.queue_id)
    )
    lobbyQueueInfo = await lobbyQueueInfoRaw.json()

    data.queue = lobbyQueueInfo["name"]
    data.queue_is_ranked = lobbyQueueInfo["isRanked"]

    module_data.rpc_updater.delay_update()


# ranked stats
@module_data.connector.ws.register(
    "/lol-ranked/v1/current-ranked-stats/", event_types=("UPDATE",)
)
async def ranked(connection: Connection, event: WebsocketEventResponse) -> None:
    data = module_data.client_data
    data.summoner_rank = (
        event.data["highestRankedEntry"]["tier"]
        + " "
        + event.data["highestRankedEntry"]["division"]
        + ": "
        + str(event.data["highestRankedEntry"]["leaguePoints"])
        + " LP"
    )

    module_data.rpc_updater.delay_update()


# ranked stats


## queue


# Base Data
# Gather base data from the LCU API on startup
async def gather_base_data(connection: Connection):
    print("Gathering base data.")
    data = module_data.client_data
    summonerDataRaw = await connection.request(
        "GET", "/lol-summoner/v1/current-summoner"
    )
    summonerData = await summonerDataRaw.json()
    data.summoner_name = summonerData["displayName"]
    data.summoner_level = summonerData["summonerLevel"]
    data.summoner_id = summonerData["summonerId"]
    data.summoner_icon = summonerData["profileIconId"]

    # get Online/Away status
    chatDataRaw = await connection.request("GET", "/lol-chat/v1/me")
    chatData = await chatDataRaw.json()
    if chatData["availability"] == "chat":
        data.availability = "Online"
    if chatData["availability"] == "away":
        data.availability = "Away"

    rankedDataRaw = await connection.request(
        "GET", "/lol-ranked/v1/current-ranked-stats/"
    )
    rankedData = await rankedDataRaw.json()
    data.summoner_rank = (
        rankedData["highestRankedEntry"]["tier"]
        + " "
        + rankedData["highestRankedEntry"]["division"]
        + ": "
        + str(rankedData["highestRankedEntry"]["leaguePoints"])
        + " LP"
    )

    gameflowDataRaw = await connection.request("GET", "/lol-gameflow/v1/gameflow-phase")
    gameflowData = await gameflowDataRaw.json()
    data.gameflow_phase = gameflowData

    if data.gameflow_phase == "InProgress":
        # In Game
        return

    if data.gameflow_phase == "None":
        # In Client
        return

    lobbyRawData = await connection.request(
        "GET", "/lol-gameflow/v1/gameflow-metadata/player-status"
    )
    lobbyData = await lobbyRawData.json()
    data.queue_id = lobbyData["currentLobbyStatus"]["queueId"]
    data.lobby_id = lobbyData["currentLobbyStatus"]["lobbyId"]
    data.players = len(lobbyData["currentLobbyStatus"]["memberSummonerIds"])

    if data.queue_id == -1:
        # custom game / practice tool / tutorial lobby
        return

    lobbyQueueInfoRaw = await connection.request(
        "GET", "/lol-game-queues/v1/queues/" + str(data.queue_id)
    )
    lobbyQueueInfo = await lobbyQueueInfoRaw.json()
    data.queue = lobbyQueueInfo["name"]
    data.max_players = int(lobbyQueueInfo["maximumParticipantListSize"])
    data.map_id = lobbyQueueInfo["mapId"]
    data.gamemode = lobbyQueueInfo["gameMode"]
    data.queue_is_ranked = lobbyQueueInfo["isRanked"]


###### Debug ######


# This will catch all events and print them to the console.
# @module_data.connector.ws.register("/", event_types=("UPDATE", "CREATE", "DELETE"))
# async def debug(connection: Connection, event: WebsocketEventResponse) -> None:
#     print(f"DEBUG - {event.type}: {event.uri}")


def update_rpc():
    print("Updating Discord Presence.")
    data = module_data.client_data
    rpc = module_data.rpc

    if data.gameflow_phase == "InProgress":
        # In Game - handled by main process
        return

    if data.gameflow_phase == "None":
        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{data.summoner_icon}.png",
            large_text="In Client",
            small_image=LEAGUE_OF_LEGENDS_LOGO,
            small_text=SMALL_TEXT,
            details="In Client",
            state=f"{data.availability}",
        )
        return

    if data.gameflow_phase == "ChampSelect":
        # In Champ Select
        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{data.summoner_icon}.png",
            large_text=f"{GAME_MODE_CONVERT_MAP.get(data.gamemode, data.gamemode)}",
            small_image=BASE_MAP_ICON_URL.format(
                map_name=MAP_ICON_CONVERT_MAP.get(data.map_id)
            ),
            small_text=SMALL_TEXT,
            details=f"In Champ Select: {data.queue}",
            state=f"Picking Champions...",
            party_size=[data.players, data.max_players],
        )
        return

    if data.gameflow_phase == "Matchmaking":
        # In Queue
        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{data.summoner_icon}.png",
            large_text=f"{GAME_MODE_CONVERT_MAP.get(data.gamemode, data.gamemode)}",
            small_image=BASE_MAP_ICON_URL.format(
                map_name=MAP_ICON_CONVERT_MAP.get(data.map_id)
            ),
            small_text=SMALL_TEXT,
            details=f"In Queue: {data.queue}",
            state=f"Searching for Game...",
            start=int(time.time()),
        )
        return

    # In Lobby

    if data.gameflow_phase == "Lobby":
        if data.is_custom or data.is_practice:
            # custom or practice tool lobby

            rpc.update(
                large_image=f"{PROFILE_ICON_BASE_URL}{str(data.summoner_icon)}.png",
                large_text=f"{GAME_MODE_CONVERT_MAP.get(data.gamemode, data.gamemode)}",
                small_image=BASE_MAP_ICON_URL.format(
                    map_name=MAP_ICON_CONVERT_MAP.get(data.map_id)
                ),
                small_text=SMALL_TEXT,
                details=f"In Lobby: {data.queue}",
                state=f"Custom Lobby",
            )

        else:
            # matchmaking lobby
            rpc.update(
                large_image=f"{PROFILE_ICON_BASE_URL}{str(data.summoner_icon)}.png",
                large_text=f"{GAME_MODE_CONVERT_MAP.get(data.gamemode, data.gamemode)}",
                small_image=BASE_MAP_ICON_URL.format(
                    map_name=MAP_ICON_CONVERT_MAP.get(data.map_id)
                ),
                small_text=SMALL_TEXT,
                details=f"In Lobby: {data.queue}",
                party_size=[data.players, data.max_players],
                state=f"Waiting for Players...",
            )
        return

    # other unhandled gameflow phases
    print(f"Unhandled Gameflow Phase: {data.gameflow_phase}")
    rpc.update(
        large_image=f"{PROFILE_ICON_BASE_URL}{str(data.summoner_icon)}.png",
        large_text=f"{data.gameflow_phase}",
        small_image=LEAGUE_OF_LEGENDS_LOGO,
        small_text=SMALL_TEXT,
        details=f"{data.gameflow_phase}",
        state="Unhandled Gameflow Phase",
    )


def start_connector(rpc_from_main: Presence) -> None:
    module_data.rpc = rpc_from_main
    module_data.update_rpc = update_rpc
    print("Starting LCU API.")
    module_data.connector.start()
