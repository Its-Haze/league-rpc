from dataclasses import dataclass
from typing import Optional


@dataclass
class ClientData:
    availability: str = "Online"  # "Online", "Away
    gamemode: Optional[str] = None
    gameflow_phase: str = "None"  # None, Lobby, Matchmaking, ReadyCheck, ChampSelect, InProgress, WaitingForStats, EndOfGame
    is_custom: bool = False
    is_practice: bool = False
    lobby_id: Optional[str] = None  # unique lobby id
    map_id: Optional[int] = None  # 11, 12, 21, 22, 30
    max_players: Optional[int] = None  # max players in lobby
    players: Optional[int] = None  # players in lobby
    queue: Optional[str] = None
    queue_id: int = -1
    queue_is_ranked: bool = False
    summoner_icon: Optional[int] = None
    summoner_id: Optional[str] = None
    summoner_level: Optional[int] = None
    summoner_name: Optional[str] = None
    summoner_rank: Optional[str] = None
