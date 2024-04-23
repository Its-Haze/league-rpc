"""
This module defines data structures for handling ranked statistics in League of Legends,
specifically tailored to interface with the /lol-ranked/v1/current-ranked-stats API endpoint. 
It provides a structured way to parse, organize, and display ranked data across different game modes
within the game, ensuring that data is accessible and usable within other components of a League client or tool.

Usage:
    These classes are essential for constructing a clear, organized view of a player's ranked status across different
    game modes in League of Legends. They facilitate the extraction and transformation of JSON data from the API
    into Python objects that can be easily manipulated and displayed, supporting features such as live display updates,
    statistical analysis, and integration with other systems or user interfaces.
"""

from dataclasses import dataclass
from typing import Any, Optional

from league_rpc.utils.const import LEAGUE_CHERRY_RANKED_EMBLEM, LEAGUE_RANKED_EMBLEM


class LolRankedRankedQueueWarnings:
    """Holds fields related to warnings about ranked queue statuses,
    such as decay warnings and demotion risks.
    """

    DAYS_UNTIL_DECAY = "daysUntilDecay"
    DEMOTION_WARNING = "demotionWarning"
    DISPLAY_DECAY_WARNING = "displayDecayWarning"
    TIME_UNTIL_INACTIVITY_STATUS_CHANGES = "timeUntilInactivityStatusChanges"


class LolRankedRankedQueueStats:
    """Contains detailed statistics and configurations for an individual ranked queue,
    including division, wins, losses, and league points.
    """

    DIVISION = "division"
    IS_PROVISIONAL = "isProvisional"
    LEAGUE_POINTS = "leaguePoints"
    LOSSES = "losses"
    MINI_SERIES_PROGRESS = "miniSeriesProgress"
    PREVIOUS_SEASON_ACHIEVED_DIVISION = "previousSeasonAchievedDivision"
    PREVIOUS_SEASON_ACHIEVED_TIER = "previousSeasonAchievedTier"
    PREVIOUS_SEASON_END_DIVISION = "previousSeasonEndDivision"
    PREVIOUS_SEASON_END_TIER = "previousSeasonEndTier"
    PROVISIONAL_GAME_THRESHOLD = "provisionalGameThreshold"
    PROVISIONAL_GAMES_REMAINING = "provisionalGamesRemaining"
    QUEUE_TYPE = "queueType"
    RATED_RATING = "ratedRating"
    RATED_TIER = "ratedTier"
    TIER = "tier"
    WARNINGS = "warnings"
    WINS = "wins"


class LolRankedRankedStats:
    """Aggregates ranked statistics across multiple queues and tracks overall progress and rewards."""

    EARNED_REGALIA_REWARD_IDS = "earnedRegaliaRewardIds"
    HIGHEST_PREVIOUS_SEASON_ACHIEVED_DIVISION = "highestPreviousSeasonAchievedDivision"
    HIGHEST_PREVIOUS_SEASON_ACHIEVED_TIER = "highestPreviousSeasonAchievedTier"
    HIGHEST_PREVIOUS_SEASON_END_DIVISION = "highestPreviousSeasonEndDivision"
    HIGHEST_PREVIOUS_SEASON_END_TIER = "highestPreviousSeasonEndTier"
    HIGHEST_RANKED_ENTRY = "highestRankedEntry"
    HIGHEST_RANKED_ENTRY_SR = "highestRankedEntrySR"
    QUEUE_MAP = "queueMap"
    QUEUES = "queues"
    RANKED_REGALIA_LEVEL = "rankedRegaliaLevel"
    SEASONS = "seasons"
    SPLITS_PROGRESS = "splitsProgress"

    # Constructor and other methods as needed


@dataclass
class RankedStats:
    """A dataclass to encapsulate specific ranked statistics for standard ranked queues,
    providing methods for easy initialization from API data and formatted string representation.
    """

    division: Optional[str] = None
    tier: Optional[str] = None
    league_points: Optional[int] = None

    @classmethod
    def from_map(
        cls,
        obj_map: dict[str, Any],
        ranked_type: str,
    ) -> "RankedStats":
        return cls(
            division=obj_map[LolRankedRankedStats.QUEUE_MAP][ranked_type][
                LolRankedRankedQueueStats.DIVISION
            ],
            tier=obj_map[LolRankedRankedStats.QUEUE_MAP][ranked_type][
                LolRankedRankedQueueStats.TIER
            ].capitalize(),
            league_points=obj_map[LolRankedRankedStats.QUEUE_MAP][ranked_type][
                LolRankedRankedQueueStats.LEAGUE_POINTS
            ],
        )

    def __str__(self) -> str:
        return f"{self.tier} {self.division}: {self.league_points} LP".strip()

    @property
    def rpc_info(self) -> tuple[str, str]:
        small_text = f"{self}"
        small_image = LEAGUE_RANKED_EMBLEM.format_map({"tier": self.tier})
        return small_text, small_image


@dataclass
class ArenaStats:
    """Similar to RankedStats, but tailored for the Arena (Project Cherry) ranked mode,
    including unique fields and mappings specific to this mode.
    """

    rated_tier: Optional[str] = None
    tier: Optional[str] = None
    rated_rating: Optional[int] = None

    tier_label_mapper = {
        "NONE": "",
        "GRAY": "Wood",
        "GREEN": "Bronze",
        "BLUE": "Silver",
        "PURPLE": "Gold",
        "ORANGE": "Gladiator",
    }

    @classmethod
    def from_map(
        cls,
        obj_map: dict[str, Any],
        ranked_type: str = "CHERRY",
    ) -> "ArenaStats":
        if not obj_map[LolRankedRankedStats.QUEUE_MAP].get(ranked_type):
            # Arena (CHERRY) is not always available to play.
            # So when it's not, we make an early return of an instance with default values.
            return cls()

        rated_tier = obj_map[LolRankedRankedStats.QUEUE_MAP][ranked_type][
            LolRankedRankedQueueStats.RATED_TIER
        ]
        return cls(
            rated_tier=rated_tier,
            tier=cls.tier_label_mapper[rated_tier],
            rated_rating=obj_map[LolRankedRankedStats.QUEUE_MAP][ranked_type][
                LolRankedRankedQueueStats.RATED_RATING
            ],
        )

    def __str__(self) -> str:
        return f"{self.tier} • Rating: {self.rated_rating}".strip()

    @property
    def rpc_info(self) -> tuple[str, str]:
        small_text = f"{self}"
        small_image = LEAGUE_CHERRY_RANKED_EMBLEM.format_map({"tier": self.tier})
        return small_text, small_image


@dataclass
class TFTStats:
    """A dataclass for encapsulating Teamfight Tactics (TFT) ranked data, structured similarly to RankedStats,
    with methods for data extraction and formatted output.
    """

    division: Optional[str] = None
    league_points: Optional[int] = None
    tier: Optional[str] = None

    @classmethod
    def from_map(
        cls,
        obj_map: dict[str, Any],
        ranked_type: str = "RANKED_TFT",
    ) -> "TFTStats":
        return cls(
            division=obj_map[LolRankedRankedStats.QUEUE_MAP][ranked_type][
                LolRankedRankedQueueStats.DIVISION
            ],
            league_points=obj_map[LolRankedRankedStats.QUEUE_MAP][ranked_type][
                LolRankedRankedQueueStats.LEAGUE_POINTS
            ],
            tier=obj_map[LolRankedRankedStats.QUEUE_MAP][ranked_type][
                LolRankedRankedQueueStats.TIER
            ].capitalize(),
        )

    def __str__(self) -> str:
        return f"{self.tier} {self.division}: {self.league_points} LP".strip()

    @property
    def rpc_info(self) -> tuple[str, str]:
        small_text = f"{self}"
        small_image = LEAGUE_RANKED_EMBLEM.format_map({"tier": self.tier})
        return small_text, small_image
