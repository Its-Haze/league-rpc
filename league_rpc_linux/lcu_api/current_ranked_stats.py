"""
This module contains the fields for extracting information from /lol-ranked/v1/current-ranked-stats
"""


class LolRankedRankedQueueWarnings:
    DAYS_UNTIL_DECAY = "daysUntilDecay"
    DEMOTION_WARNING = "demotionWarning"
    DISPLAY_DECAY_WARNING = "displayDecayWarning"
    TIME_UNTIL_INACTIVITY_STATUS_CHANGES = "timeUntilInactivityStatusChanges"


class LolRankedRankedQueueStats:
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
