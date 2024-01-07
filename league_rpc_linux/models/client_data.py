import time
from dataclasses import dataclass, field

from league_rpc_linux.models.lcu.current_ranked_stats import (
    ArenaStats,
    RankedStats,
    TFTStats,
)


@dataclass
class ClientData:
    availability: str = "Online"  # "Online", "Away
    gamemode: str = ""
    gameflow_phase: str = "None"  # None, Lobby, Matchmaking, ReadyCheck, ChampSelect, InProgress, WaitingForStats, EndOfGame
    is_custom: bool = False
    is_practice: bool = False
    lobby_id: str = ""  # unique lobby id
    map_id: int = 0  # 11, 12, 21, 22, 30
    max_players: int = 0  # max players in lobby
    players: int = 0  # players in lobby
    queue: str = ""
    queue_type = ""
    queue_id: int = -1
    queue_is_ranked: bool = False
    summoner_icon: int = 0
    summoner_id: str = ""
    summoner_level: int = 0
    summoner_name: str = ""
    summoner_rank: RankedStats = field(default_factory=RankedStats)
    summoner_rank_flex: RankedStats = field(default_factory=RankedStats)
    arena_rank: ArenaStats = field(default_factory=ArenaStats)
    tft_rank: TFTStats = field(default_factory=TFTStats)
    application_start_time: int = int(time.time())
