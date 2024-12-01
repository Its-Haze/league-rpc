"""
This module defines the ClientData class, a data structure designed to hold comprehensive
client-related information about a player's session in League of Legends. This class is particularly
useful for tracking real-time client state and player statistics, encapsulating everything from the
current game mode to the player's ranked stats across different game formats.


Usage:
    The ClientData class is integral to systems that require monitoring or displaying real-time information
    about a player’s current game status or session details. It is especially useful in client applications,
    overlays, and tools that provide enhanced player experiences by utilizing live data to offer insights,
    stats tracking, and session analysis.
"""

import time
from dataclasses import dataclass, field

from league_rpc.models.lcu.current_ranked_stats import ArenaStats, RankedStats, TFTStats


@dataclass
class ClientData:
    """Stores data relevant to the player's current session, including their availability,
    game mode, lobby details, and ranked statistics. The class uses the dataclass decorator
    for convenient storage and retrieval of instance data with default values and type annotations.
    """

    availability: str = "Online"  # "Online", "Away
    gamemode: str = ""
    gameflow_phase: str = "None"  # None, Lobby, Matchmaking, ReadyCheck, ChampSelect, InProgress, WaitingForStats, EndOfGame
    is_custom: bool = False
    is_practice: bool = False
    lobby_id: str = ""  # unique lobby id
    map_id: int = 0  # 11, 12, 21, 22, 30
    max_players: int = 0  # max players in lobby
    players: int = 0  # players in lobby
    queue_name: str = ""
    queue_type: str = ""
    queue_is_ranked: bool = False
    queue_detailed_description: str = ""
    queue_description: str = ""
    queue_id: int = -1
    summoner_icon: int = 0
    summoner_rank: RankedStats = field(default_factory=RankedStats)
    summoner_rank_flex: RankedStats = field(default_factory=RankedStats)
    arena_rank: ArenaStats = field(default_factory=ArenaStats)
    tft_rank: TFTStats = field(default_factory=TFTStats)
    tft_companion_id: int = 0
    tft_companion_icon: str = ""
    tft_companion_name: str = ""
    tft_companion_description: str = ""
    application_start_time: int = int(time.time())

    @property
    def get_queue_name(self) -> str:
        """Return the detailed descriptive name of the queue if available, otherwise return the queue name."""
        return (
            self.queue_detailed_description
            if self.queue_detailed_description
            else self.queue_name
        )
