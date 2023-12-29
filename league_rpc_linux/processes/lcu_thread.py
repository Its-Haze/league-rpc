import time
from threading import Timer

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

connector = lcu_driver.Connector()

rpc = None

data = {
    "champion_name": None,
    "skin_name": None,
    "skin_id": None,
    "level": None,
    "gamemode": None,
    "isCustom": False,
    "isPractice": False,
    "inLobby": False,
    "inQueue": False,
    "inChampSelect": False,
    "max_players": None,
    "players": None,
    "summoner_name": None,
    "summoner_id": None,
    "summoner_icon": None,
    "summoner_rank": None,
    "queue": None,
    "queue_id": -1,
    "map_id": None,  # 11, 12, 21, 22, 30
    "lobbyQueueIsRanked": False,  # True, False
    "lobby_id": None,
    "availability": "Idle",  # Idle, Away
    "gameflowPhase": None,  # Lobby, Matchmaking, ReadyCheck, ChampSelect, InProgress, WaitingForStats, EndOfGame
}

## Timers ##

scheduledUpdate: bool = False


# As some events are called multiple times, we should limit the amount of updates to the RPC.
# Collect update events for 1 second and then update the RPC.
def delayUpdate():
    global scheduledUpdate

    if not scheduledUpdate:
        scheduledUpdate = True
        Timer(1.0, update_rpc_and_reset_flag).start()


def update_rpc_and_reset_flag():
    global scheduledUpdate
    update_rpc()
    scheduledUpdate = False


## WS Events ##


@connector.ready
async def connect(connection: Connection):
    print(f"{Colors.green}LCU API is ready.{Colors.reset}")

    await gather_base_data(connection)
    delayUpdate()


@connector.close
async def disconnect(connection: Connection):
    print("LCU API is closed.")


@connector.ws.register("/lol-summoner/v1/current-summoner", event_types=("UPDATE",))
async def summoner_updated(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    print("Summoner has been updated.")
    data["summoner_name"] = event.data["displayName"]
    data["level"] = event.data["summonerLevel"]
    data["summoner_id"] = event.data["summonerId"]
    data["summoner_icon"] = event.data["profileIconId"]

    delayUpdate()


@connector.ws.register("/lol-chat/v1/me", event_types=("UPDATE",))
async def chat_updated(connection: Connection, event: WebsocketEventResponse) -> None:
    print("Chat has been updated.")
    if event.data["availability"] == "chat":
        data["availability"] = "Idle"
    if event.data["availability"] == "away":
        data["availability"] = "Away"

    delayUpdate()


@connector.ws.register("/lol-gameflow/v1/gameflow-phase", event_types=("UPDATE",))
async def gameflow_phase_updated(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    print("Gameflow Phase has been updated.")
    data["gameflowPhase"] = event.data  # returns plain string of the phase

    delayUpdate()


@connector.ws.register("/lol-champ-select/v1/session", event_types=("CREATE", "DELETE"))
async def champ_select_joined(
    connection: Connection, event: WebsocketEventResponse
) -> None:
    if event.type == "Create":
        data["inChampSelect"] = True
        delayUpdate()
        return

    if event.type == "Delete":
        data["inChampSelect"] = False
        delayUpdate()
        return


# could be used for lobby instead: /lol-gameflow/v1/gameflow-metadata/player-status
@connector.ws.register(
    "/lol-lobby/v2/lobby", event_types=("UPDATE", "CREATE", "DELETE")
)
async def in_lobby(connection: Connection, event: WebsocketEventResponse) -> None:
    print(f"Lobby Data - {event.type}")

    if event.type == "Delete":
        data["inLobby"] = False
        delayUpdate()
        return

    data["inLobby"] = True

    if event.type == "Create":
        # wait for Update Call, as create does not contain data
        return

    data["queue_id"] = event.data["gameConfig"]["queueId"]
    data["lobby_id"] = event.data["partyId"]
    data["players"] = len(event.data["members"])
    data["max_players"] = int(event.data["gameConfig"]["maxLobbySize"])
    data["map_id"] = event.data["gameConfig"]["mapId"]
    data["gamemode"] = event.data["gameConfig"]["gameMode"]
    data["isCustom"] = event.data["gameConfig"]["isCustom"]
    if event.data["gameConfig"]["gameMode"] == "PRACTICETOOL":
        data["isPractice"] = True
        data["max_players"] = 1

    if data["queue_id"] == -1:
        # custom game / practice tool / tutorial lobby
        if data["isPractice"]:
            data["queue"] = "Practice Tool"
        else:
            data["queue"] = "Custom Game"
        delayUpdate()
        return

    lobbyQueueInfoRaw = await connection.request(
        "GET", "/lol-game-queues/v1/queues/" + str(data["queue_id"])
    )
    lobbyQueueInfo = await lobbyQueueInfoRaw.json()

    data["queue"] = lobbyQueueInfo["name"]
    data["lobbyQueueIsRanked"] = lobbyQueueInfo["isRanked"]

    delayUpdate()


# ranked stats
@connector.ws.register("/lol-ranked/v1/current-ranked-stats/", event_types=("UPDATE",))
async def ranked(connection: Connection, event: WebsocketEventResponse) -> None:
    data["summoner_rank"] = (
        event.data["highestRankedEntry"]["tier"]
        + " "
        + event.data["highestRankedEntry"]["division"]
        + ": "
        + str(event.data["highestRankedEntry"]["leaguePoints"])
        + " LP"
    )

    delayUpdate()


# ranked stats


## queue
@connector.ws.register(
    "/lol-matchmaking/v1/search",
    event_types=(
        "CREATE",
        "DELETE",
    ),
)
async def matchmaking(connection: Connection, event: WebsocketEventResponse) -> None:
    print(f"Matchmaking - {event.type}")
    if event.type == "Create":
        data["inQueue"] = True
    if event.type == "Delete":
        data["inQueue"] = False
    delayUpdate()


## queue


# Base Data
# Gather base data from the LCU API on startup
async def gather_base_data(connection: Connection):
    print("Gathering base data.")
    summonerDataRaw = await connection.request(
        "GET", "/lol-summoner/v1/current-summoner"
    )
    summonerData = await summonerDataRaw.json()
    data["summoner_name"] = summonerData["displayName"]
    data["level"] = summonerData["summonerLevel"]
    data["summoner_id"] = summonerData["summonerId"]
    data["summoner_icon"] = summonerData["profileIconId"]

    # get Idle/Away status
    chatDataRaw = await connection.request("GET", "/lol-chat/v1/me")
    chatData = await chatDataRaw.json()
    if chatData["availability"] == "chat":
        data["availability"] = "Online"
    if chatData["availability"] == "away":
        data["availability"] = "Away"

    rankedDataRaw = await connection.request(
        "GET", "/lol-ranked/v1/current-ranked-stats/"
    )
    rankedData = await rankedDataRaw.json()
    data["summoner_rank"] = (
        rankedData["highestRankedEntry"]["tier"]
        + " "
        + rankedData["highestRankedEntry"]["division"]
        + ": "
        + str(rankedData["highestRankedEntry"]["leaguePoints"])
        + " LP"
    )

    lobbyRawData = await connection.request(
        "GET", "/lol-gameflow/v1/gameflow-metadata/player-status"
    )
    lobbyData = await lobbyRawData.json()
    data["queue_id"] = lobbyData["currentLobbyStatus"]["queueId"]
    data["lobby_id"] = lobbyData["currentLobbyStatus"]["lobbyId"]
    data["players"] = len(lobbyData["currentLobbyStatus"]["memberSummonerIds"])

    if data["queue_id"] == -1:
        # Not in Lobby
        return

    data["inLobby"] = True
    lobbyQueueInfoRaw = await connection.request(
        "GET", "/lol-game-queues/v1/queues/" + str(data["queue_id"])
    )
    lobbyQueueInfo = await lobbyQueueInfoRaw.json()
    data["queue"] = lobbyQueueInfo["name"]
    data["max_players"] = int(lobbyQueueInfo["maximumParticipantListSize"])
    data["map_id"] = lobbyQueueInfo["mapId"]
    data["gamemode"] = lobbyQueueInfo["gameMode"]
    data["lobbyQueueIsRanked"] = lobbyQueueInfo["isRanked"]


###### Debug ######


# This will catch all events and print them to the console.
# @connector.ws.register("/", event_types=("UPDATE", "CREATE", "DELETE"))
# async def debug(connection: Connection, event: WebsocketEventResponse) -> None:
#     print(f"DEBUG - {event.type}: {event.uri}")


def update_rpc():
    print("Updating Discord Presence.")

    if data["gameflowPhase"] == "InProgress":
        # In Game - handled by main process
        return

    if not data["inLobby"]:
        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{data['summoner_icon']}.png",
            large_text="In Client",
            small_image=LEAGUE_OF_LEGENDS_LOGO,
            small_text=SMALL_TEXT,
            details="In Client",
            state=f"{data['availability']}",
        )
        return

    if data["inChampSelect"]:
        # In Champ Select
        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{data['summoner_icon']}.png",
            large_text=f"{GAME_MODE_CONVERT_MAP.get(data['gamemode'], data['gamemode'])}",
            small_image=BASE_MAP_ICON_URL.format(
                map_name=MAP_ICON_CONVERT_MAP.get(data["map_id"])
            ),
            small_text=SMALL_TEXT,
            details=f"In Champ Select: {data['queue']}",
            state=f"Picking Champions...",
            party_size=[data["players"], data["max_players"]],
        )
        return

    if data["inQueue"]:
        # In Queue
        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{data['summoner_icon']}.png",
            large_text=f"{GAME_MODE_CONVERT_MAP.get(data['gamemode'], data['gamemode'])}",
            small_image=BASE_MAP_ICON_URL.format(
                map_name=MAP_ICON_CONVERT_MAP.get(data["map_id"])
            ),
            small_text=SMALL_TEXT,
            details=f"In Queue: {data['queue']}",
            state=f"Searching for Game...",
            start=int(time.time()),
        )
        return

    # In Lobby

    if data["queue_id"] == -1:
        # custom or practice tool lobby

        rpc.update(
            large_image=f"{PROFILE_ICON_BASE_URL}{str(data['summoner_icon'])}.png",
            large_text=f"{GAME_MODE_CONVERT_MAP.get(data['gamemode'], data['gamemode'])}",
            small_image=BASE_MAP_ICON_URL.format(
                map_name=MAP_ICON_CONVERT_MAP.get(data["map_id"])
            ),
            small_text=SMALL_TEXT,
            details=f"In Lobby: {data['queue']}",
            state=f"Custom Lobby",
        )

    # matchmaking lobby
    rpc.update(
        large_image=f"{PROFILE_ICON_BASE_URL}{str(data['summoner_icon'])}.png",
        large_text=f"{GAME_MODE_CONVERT_MAP.get(data['gamemode'], data['gamemode'])}",
        small_image=BASE_MAP_ICON_URL.format(
            map_name=MAP_ICON_CONVERT_MAP.get(data["map_id"])
        ),
        small_text=SMALL_TEXT,
        details=f"In Lobby: {data['queue']}",
        party_size=[data["players"], data["max_players"]],
        state=f"Waiting for Players...",
    )


def start_connector(rpc_from_main: Presence):
    global rpc
    rpc = rpc_from_main
    print("Starting LCU API.")
    connector.start()
