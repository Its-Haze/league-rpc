import time
import lcu_driver
from league_rpc_linux.colors import Colors

from league_rpc_linux.const import (
    BASE_MAP_URL,
    BASE_CHAMPION_URL,
    BASE_SKIN_URL,
    GAME_MODE_CONVERT_MAP,
    LEAGUE_OF_LEGENDS_LOGO,
)


connector = lcu_driver.Connector()

rpc = None

data = {
    "champion_name": None,
    "skin_name": None,
    "skin_id": None,
    "game_mode": None,
    "level": None,
    "queue": None,
    "queueState": None,
    "gamemode": None,
    "champSelect": None,
    "max_players": None,
    "players": None,
    "summoner_name": None,
    "summoner_id": None,
    "summoner_rank": None,
    "queue_id": -1,
    "map_id": None,
    "lobbyQueueIsRanked": None,
    "lobby_id": None,
}


@connector.ready
async def connect(connection):
    print(f"{Colors.green}LCU API is ready.{Colors.reset}")

    await gather_based_data(connection)
    await update_rpc()


@connector.close
async def disconnect(connection):
    print("LCU API is closed.")


@connector.ws.register("/lol-champ-select/v1/session", event_types=("UPDATE",))
async def champ_select(connection, event):
    print("Champ select has been updated.")


@connector.ws.register(
    "/lol-gameflow/v1/gameflow-metadata/player-status", event_types=("UPDATE",)
)
async def in_lobby(connection, event):
    print("In lobby has been updated.")

    # Artificially set queueState to None - if the /search request is 404, the websocket event will not be triggered. This is a workaround.
    data["queueState"] = None

    data["queue_id"] = event.data["currentLobbyStatus"]["queueId"]
    data["lobby_id"] = event.data["currentLobbyStatus"]["lobbyId"]
    data["players"] = len(event.data["currentLobbyStatus"]["memberSummonerIds"])

    if data["queue_id"] == -1:
        # updated lobby data. not in lobby anymore
        await update_rpc()
        return

    lobbyQueueInfoRaw = await connection.request(
        "GET", "/lol-game-queues/v1/queues/" + str(data["queue_id"])
    )
    lobbyQueueInfo = await lobbyQueueInfoRaw.json()

    data["queue"] = lobbyQueueInfo["name"]
    data["max_players"] = lobbyQueueInfo["maximumParticipantListSize"]
    data["map_id"] = lobbyQueueInfo["mapId"]
    data["gamemode"] = lobbyQueueInfo["gameMode"]
    data["lobbyQueueIsRanked"] = lobbyQueueInfo["isRanked"]

    await update_rpc()


@connector.ws.register("/lol-ranked/v1/current-ranked-stats/", event_types=("UPDATE",))
async def ranked(connection, event):
    data["summoner_rank"] = (
        event.data["highestRankedEntry"]["tier"]
        + " "
        + event.data["highestRankedEntry"]["division"]
        + ": "
        + str(event.data["highestRankedEntry"]["leaguePoints"])
        + " LP"
    )

    await update_rpc()


@connector.ws.register("/lol-matchmaking/v1/search", event_types=("UPDATE",))
async def matchmaking(connection, event):
    dataSearchState = event.data.get("searchState", None)

    if data["queueState"] == dataSearchState:
        # No change in state, only timer updated.
        # print(f"Debug: {dataSearchState} == {data['queueState']}")
        return
    data["queueState"] = event.data.get("searchState", None)

    print("Matchmaking has been updated.")
    await update_rpc()


async def gather_based_data(connection):
    print("Gathering base data.")
    summonerDataRaw = await connection.request(
        "GET", "/lol-summoner/v1/current-summoner"
    )
    summonerData = await summonerDataRaw.json()
    data["summoner_name"] = summonerData["displayName"]
    data["level"] = summonerData["summonerLevel"]
    data["summoner_id"] = summonerData["summonerId"]

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

    lobbyQueueInfoRaw = await connection.request(
        "GET", "/lol-game-queues/v1/queues/" + str(data["queue_id"])
    )
    lobbyQueueInfo = await lobbyQueueInfoRaw.json()
    data["queue"] = lobbyQueueInfo["name"]
    data["max_players"] = lobbyQueueInfo["maximumParticipantListSize"]
    data["map_id"] = lobbyQueueInfo["mapId"]
    data["gamemode"] = lobbyQueueInfo["gameMode"]
    data["lobbyQueueIsRanked"] = lobbyQueueInfo["isRanked"]


async def update_rpc():
    print("Updating Discord Presence.")
    if data["queue_id"] == -1:
        # Not in Lobby
        rpc.update(
            large_image=LEAGUE_OF_LEGENDS_LOGO,
            large_text="In Client",
            details="In Client",
            state=f"Level {data['level']}",
        )
        return

    if data["queueState"] == "Searching":
        # In Queue
        rpc.update(
            large_image=f"{BASE_MAP_URL}{str(data['map_id'])}.png",
            large_text=f"{GAME_MODE_CONVERT_MAP.get(data['gamemode'], data['gamemode'])}",
            details=f"In Queue: {data['queue']}",
            state=f"Level {data['level']}",
            start=int(time.time()),
        )
        return

    # In Lobby
    rpc.update(
        large_image=f"{BASE_MAP_URL}{str(data['map_id'])}.png",
        large_text=f"{GAME_MODE_CONVERT_MAP.get(data['gamemode'], data['gamemode'])}",
        details=f"In Lobby: {data['queue']}",
        party_size=[data["players"], data["max_players"]],
        state=f"Level {data['level']}",
    )


def start_connector(rpc_from_main):
    global rpc
    rpc = rpc_from_main
    print("Starting LCU API.")
    connector.start()
