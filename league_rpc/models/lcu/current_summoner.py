"""
This module defines a class that encapsulates information about a summoner in League of Legends,
specifically for interfacing with the /lol-summoner/v1/current-summoner API endpoint. It structures
the data associated with a summoner's account, providing an organized way to access properties such
as display name, level, profile icon, and experience points.

Usage:
    This class serves as a centralized resource for retrieving and manipulating summoner data within
    applications. It is particularly useful in League of Legends client tools, where displaying up-to-date
    summoner information or interacting with the client's features requires accurate and organized data.
    The fields in this class can be used to map JSON data from the API into Python objects for further processing
    or display, facilitating easy access and manipulation of summoner-specific information.
"""


class Summoner:
    """Contains constants for key fields extracted from the summoner data endpoint,
    including identifiers, level information, and details on customization points (reroll points)
    used in certain game modes like ARAM (All Random All Mid).
    """

    ACCOUNT_ID = "accountId"
    DISPLAY_NAME = "displayName"
    INTERNAL_NAME = "internalName"
    NAME_CHANGE_FLAG = "nameChangeFlag"
    PERCENT_COMPLETE_FOR_NEXT_LEVEL = "percentCompleteForNextLevel"
    PRIVACY = "privacy"
    PROFILE_ICON_ID = "profileIconId"
    PUUID = "puuid"
    REROLL_POINTS = "rerollPoints"
    CURRENT_POINTS = "currentPoints"
    MAX_ROLL = "maxRolls"
    NUMBER_OF_ROLLS = "numberOfRolls"
    POINTS_COST_TO_ROLL = "pointsCostToRoll"
    SUMMONER_ID = "summonerId"
    SUMMONER_LEVEL = "summonerLevel"
    UNNAMED = "unnamed"
    XP_SINCE_LAST_LEVEL = "xpSinceLastLevel"
    XP_UNTIL_NEXT_LEVEL = "xpUntilNextLevel"
