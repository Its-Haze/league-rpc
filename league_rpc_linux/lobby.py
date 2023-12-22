from league_rpc_linux.lcu.LCU import getLcuAPI

lcu = getLcuAPI()


def checkLcuAPI():
    print(lcu.getDataFromPath("/lol-summoner/v1/current-summoner").items())


def inClient():
    lobbyStatus = lcu.getDataFromPath(
        "/lol-gameflow/v1/gameflow-metadata/player-status"
    )

    if not lobbyStatus:
        return False

    # Not in Lobby
    if lobbyStatus["currentLobbyStatus"]["queueId"] == -1:
        summonerData = lcu.getDataFromPath("/lol-summoner/v1/current-summoner")
        return {
            "type": "summoner",
            "level": summonerData["summonerLevel"],
        }

    # In Lobby
    queueId = lobbyStatus["currentLobbyStatus"]["queueId"]

    lobbyQueueInfo = lcu.getDataFromPath("/lol-game-queues/v1/queues/" + str(queueId))

    return {
        "type": "lobby",
        "queue": lobbyQueueInfo["description"],
        "maxPlayers": lobbyQueueInfo["maximumParticipantListSize"],
        "players": len(lobbyStatus["currentLobbyStatus"]["memberSummonerIds"]),
    }
