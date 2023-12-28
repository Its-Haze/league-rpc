import time
from threading import Timer

import lcu_driver
from lcu_driver.connection import Connection
from lcu_driver.events.responses import WebsocketEventResponse

from league_rpc_linux.colors import Colors
from league_rpc_linux.const import (
    BASE_MAP_ICON_URL,
    GAME_MODE_CONVERT_MAP,
    LEAGUE_OF_LEGENDS_LOGO,
    MAP_ICON_CONVERT_MAP,
    PROFILE_ICON_BASE_URL,
)

## Timers ##


# As some events are called multiple times, we should limit the amount of updates to the RPC.
# Collect update events for 1 second and then update the RPC.
class UpdateScheduler:
    def __init__(self, lcu) -> None:
        self.lcu = lcu
        self.update_already_scheduled: bool = False

    def delay_update(self) -> None:
        if not self.update_already_scheduled:
            self.update_already_scheduled = True
            Timer(1.0, self.update_rpc_and_reset_flag).start()

    def update_rpc_and_reset_flag(self) -> None:
        self.lcu.update_rpc()
        self.update_already_scheduled = False


## WS Events ##


class LcuThread:
    def __init__(self, rpc_from_main) -> None:
        self.rpc = rpc_from_main
        self.connector = lcu_driver.Connector()

        self.update_scheduler = UpdateScheduler(self)
        self.data = {
            "champion_name": None,
            "skin_name": None,
            "skin_id": None,
            "level": None,
            "gamemode": None,
            "is_custom": False,
            "is_practice": False,
            "in_lobby": False,
            "in_queue": False,
            "in_champ_select": False,
            "max_players": None,
            "players": None,
            "summoner_name": None,
            "summoner_id": None,
            "summoner_icon": None,
            "summoner_rank": None,
            "queue": None,
            "queue_id": -1,
            "map_id": None,  # 11, 12, 21, 22, 30
            "is_ranked_queue": False,  # True, False
            "lobby_id": None,
            "availability": "Idle",  # Idle, Away
            "gamflow_phase": None,  # Lobby, Matchmaking, ReadyCheck, ChampSelect, InProgress, WaitingForStats, EndOfGame
        }

        self.register_ws_events()
        print("Starting connector to LCU API.")
        self.connector.start()

    # Register WS Events
    # The Decorator will not work inside classes, thus we need to explicitly register the events manually.
    def register_ws_events(self) -> None:
        self.connector.ready(self.connect)
        self.connector.close(self.disconnect)
        self.connector.ws.register(
            "/lol-summoner/v1/current-summoner",
            event_types=("UPDATE",),
        )(self.summoner_updated)
        self.connector.ws.register("/lol-chat/v1/me", event_types=("UPDATE",))(
            self.chat_updated
        )
        self.connector.ws.register(
            "/lol-gameflow/v1/gameflow-phase",
            event_types=("UPDATE",),
        )(self.gameflow_phase_updated)
        self.connector.ws.register(
            "/lol-champ-select/v1/session",
            event_types=("CREATE", "DELETE"),
        )(self.champ_select)
        self.connector.ws.register(
            "/lol-lobby/v2/lobby",
            event_types=("UPDATE", "CREATE", "DELETE"),
        )(self.in_lobby)
        self.connector.ws.register(
            "/lol-ranked/v1/current-ranked-stats/",
            event_types=("UPDATE",),
        )(self.ranked)
        self.connector.ws.register(
            "/lol-matchmaking/v1/search",
            event_types=("DELETE", "CREATE"),
        )(self.matchmaking)
        # self.connector.ws.register('/', event_types=('UPDATE','CREATE', 'REMOVE'))(self.debug)

    async def connect(self, connection: Connection) -> None:
        print(f"{Colors.green}LCU API is ready.{Colors.reset}")

        await self.gather_base_data(connection)
        self.update_scheduler.delay_update()

    async def disconnect(self, connection: Connection) -> None:
        print("LCU API is closed.")

    async def summoner_updated(
        self,
        connection: Connection,
        event: WebsocketEventResponse,
    ) -> None:
        print("Summoner has been updated.")
        self.data["summoner_name"] = event.data["displayName"]
        self.data["level"] = event.data["summonerLevel"]
        self.data["summoner_id"] = event.data["summonerId"]
        self.data["summoner_icon"] = event.data["profileIconId"]

        self.update_scheduler.delay_update()

    async def chat_updated(
        self,
        connection: Connection,
        event: WebsocketEventResponse,
    ):
        print("Chat has been updated.")
        if event.data["availability"] == "chat":
            self.data["availability"] = "Online"
        if event.data["availability"] == "away":
            self.data["availability"] = "Away"

        self.update_scheduler.delay_update()

    async def gameflow_phase_updated(
        self,
        connection: Connection,
        event: WebsocketEventResponse,
    ) -> None:
        print("Gameflow Phase has been updated.")
        self.data["gamflow_phase"] = event.data  # returns plain string of the phase

        self.update_scheduler.delay_update()

    async def champ_select(
        self,
        connection: Connection,
        event: WebsocketEventResponse,
    ) -> None:
        print(f"Champ Select data - {event.type}")
        if event.type == "DELETE":
            self.data["in_champ_select"] = False

        if event.type == "CREATE":
            self.data["in_champ_select"] = True

        self.update_scheduler.delay_update()

    async def in_lobby(
        self,
        connection: Connection,
        event: WebsocketEventResponse,
    ) -> None:
        print(f"Lobby data - {event.type}")

        if event.type == "Delete":
            self.data["in_lobby"] = False
            self.update_scheduler.delay_update()
            return

        self.data["in_lobby"] = True

        if event.type == "Create":
            # wait for Update Call, as create does not contain event.data
            return

        self.data["queue_id"] = event.data["gameConfig"]["queueId"]
        self.data["lobby_id"] = event.data["partyId"]
        self.data["players"] = len(event.data["members"])
        self.data["max_players"] = int(event.data["gameConfig"]["maxLobbySize"])
        self.data["map_id"] = event.data["gameConfig"]["mapId"]
        self.data["gamemode"] = event.data["gameConfig"]["gameMode"]
        self.data["is_custom"] = event.data["gameConfig"]["isCustom"]
        if event.data["gameConfig"]["gameMode"] == "PRACTICETOOL":
            self.data["is_practice"] = True
            self.data["max_players"] = 1

        if self.data["queue_id"] == -1:
            # custom game / practice tool / tutorial lobby
            if self.data["is_practice"]:
                self.data["queue"] = "Practice Tool"
            else:
                self.data["queue"] = "Custom Game"
            self.update_scheduler.delay_update()
            return

        lobbyQueueInfoRaw = await connection.request(
            "GET", "/lol-game-queues/v1/queues/" + str(self.data["queue_id"])
        )
        lobbyQueueInfo = await lobbyQueueInfoRaw.json()

        self.data["queue"] = lobbyQueueInfo["name"]
        self.data["is_ranked_queue"] = lobbyQueueInfo["isRanked"]

        self.update_scheduler.delay_update()

    # ranked stats
    async def ranked(
        self,
        connection: Connection,
        event: WebsocketEventResponse,
    ) -> None:
        self.data["summoner_rank"] = (
            event.data["highestRankedEntry"]["tier"]
            + " "
            + event.data["highestRankedEntry"]["division"]
            + ": "
            + str(event.data["highestRankedEntry"]["leaguePoints"])
            + " LP"
        )

        self.update_scheduler.delay_update()

    # ranked stats

    ## queue
    async def matchmaking(
        self,
        connection: Connection,
        event: WebsocketEventResponse,
    ) -> None:
        if event.type == "DELETE":
            self.data["in_queue"] = False
            self.update_scheduler.delay_update()
            return

        self.data["in_queue"] = True
        self.update_scheduler.delay_update()

    ###### Debug ######

    # This will catch all events and print them to the console.
    # async def debug(connection: Connection, event: WebsocketEventResponse):
    #     print(f"DEBUG - {event.type}: {event.uri}")

    # Base data
    # Gather base data from the LCU API on startup
    async def gather_base_data(self, connection: Connection) -> None:
        print("Gathering base data.")
        summonerDataRaw = await connection.request(
            "GET", "/lol-summoner/v1/current-summoner"
        )
        summonerData = await summonerDataRaw.json()
        self.data["summoner_name"] = summonerData["displayName"]
        self.data["level"] = summonerData["summonerLevel"]
        self.data["summoner_id"] = summonerData["summonerId"]
        self.data["summoner_icon"] = summonerData["profileIconId"]

        # get Idle/Away status
        chatDataRaw = await connection.request("GET", "/lol-chat/v1/me")
        chatData = await chatDataRaw.json()
        if chatData["availability"] == "chat":
            self.data["availability"] = "Online"
        if chatData["availability"] == "away":
            self.data["availability"] = "Away"

        rankedDataRaw = await connection.request(
            "GET", "/lol-ranked/v1/current-ranked-stats/"
        )
        rankedData = await rankedDataRaw.json()
        self.data["summoner_rank"] = (
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
        self.data["queue_id"] = lobbyData["currentLobbyStatus"]["queueId"]
        self.data["lobby_id"] = lobbyData["currentLobbyStatus"]["lobbyId"]
        self.data["players"] = len(lobbyData["currentLobbyStatus"]["memberSummonerIds"])

        if self.data["queue_id"] == -1:
            # Not in Lobby
            return

        self.data["in_lobby"] = True
        lobbyQueueInfoRaw = await connection.request(
            "GET", "/lol-game-queues/v1/queues/" + str(self.data["queue_id"])
        )
        lobbyQueueInfo = await lobbyQueueInfoRaw.json()
        self.data["queue"] = lobbyQueueInfo["name"]
        self.data["max_players"] = int(lobbyQueueInfo["maximumParticipantListSize"])
        self.data["map_id"] = lobbyQueueInfo["mapId"]
        self.data["gamemode"] = lobbyQueueInfo["gameMode"]
        self.data["is_ranked_queue"] = lobbyQueueInfo["isRanked"]

    def update_rpc(self) -> None:
        print("Updating Discord Presence.")

        if self.data["gamflow_phase"] == "InProgress":
            # In Game - handled by main process
            return

        if not self.data["in_lobby"]:
            self.rpc.update(
                large_image=f"{PROFILE_ICON_BASE_URL}{str(self.data['summoner_icon'])}.png",
                large_text="In Client",
                small_image=LEAGUE_OF_LEGENDS_LOGO,
                small_text="League of Legends Logo",
                details="In Client",
                state=f"{self.data['availability']}",
            )
            return

        if self.data["in_champ_select"]:
            # In Champ Select
            self.rpc.update(
                large_image=BASE_MAP_ICON_URL.format(
                    map_name=MAP_ICON_CONVERT_MAP.get(self.data["map_id"])
                ),
                large_text=f"{GAME_MODE_CONVERT_MAP.get(self.data['gamemode'], self.data['gamemode'])}",
                small_image=LEAGUE_OF_LEGENDS_LOGO,
                small_text="League of Legends Logo",
                details=f"In Champ Select: {self.data['queue']}",
                state=f"Picking Champions...",
                party_size=[self.data["players"], self.data["max_players"]],
            )
            return

        if self.data["in_queue"]:
            # In Queue
            self.rpc.update(
                large_image=BASE_MAP_ICON_URL.format(
                    map_name=MAP_ICON_CONVERT_MAP.get(self.data["map_id"])
                ),
                large_text=f"{GAME_MODE_CONVERT_MAP.get(self.data['gamemode'], self.data['gamemode'])}",
                small_image=LEAGUE_OF_LEGENDS_LOGO,
                small_text="League of Legends Logo",
                details=f"In Queue: {self.data['queue']}",
                state="Searching for Game...",
                start=int(time.time()),
            )
            return

        # In Lobby

        if self.data["queue_id"] == -1:
            # custom or practice tool lobby

            self.rpc.update(
                large_image=BASE_MAP_ICON_URL.format(
                    map_name=MAP_ICON_CONVERT_MAP.get(self.data["map_id"])
                ),
                large_text=f"{GAME_MODE_CONVERT_MAP.get(self.data['gamemode'], self.data['gamemode'])}",
                small_image=LEAGUE_OF_LEGENDS_LOGO,
                small_text="League of Legends Logo",
                details=f"In Lobby: {self.data['queue']}",
                state="Custom Lobby",
            )

        # matchmaking lobby
        self.rpc.update(
            large_image=BASE_MAP_ICON_URL.format(
                map_name=MAP_ICON_CONVERT_MAP.get(self.data["map_id"])
            ),
            large_text=f"{GAME_MODE_CONVERT_MAP.get(self.data['gamemode'], self.data['gamemode'])}",
            small_image=LEAGUE_OF_LEGENDS_LOGO,
            small_text="League of Legends Logo",
            details=f"In Lobby: {self.data['queue']}",
            party_size=[self.data["players"], self.data["max_players"]],
            state="Waiting for Players...",
        )


def start_connector(rpc_from_main):
    LcuThread(rpc_from_main)
