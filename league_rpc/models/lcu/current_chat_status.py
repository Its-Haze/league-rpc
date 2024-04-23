"""
This module contains the fields for extracting information from /lol-chat/v1/me
"""


class LolChatUser:
    """
    Represents a user in the League of Legends (LoL) chat system, encapsulating all relevant attributes
    pertaining to a user's profile within the game's chat interface. This class serves as a structured way
    to define and access various constants related to a user's chat properties, such as availability,
    game-related information, identity, and status details.

    Usage:
        This class does not encapsulate behaviors and is utilized primarily for defining and using string constants in a manner that prevents typos and errors in key usage across the chat system implementation.
    """

    AVAILABILITY = "availability"
    GAME_NAME = "gameName"
    GAME_TAG = "gameTag"
    ICON = "icon"
    ID = "id"
    LAST_SEEN_ONLINE_TIMESTAMP = "lastSeenOnlineTimestamp"
    LOL = "lol"  # This is a dictionary with unspecified key-value pairs
    NAME = "name"
    PATCHLINE = "patchline"
    PID = "pid"
    PLATFORM_ID = "platformId"
    PRODUCT = "product"
    PRODUCT_NAME = "productName"
    PUUID = "puuid"
    STATUS_MESSAGE = "statusMessage"
    SUMMARY = "summary"
    SUMMONER_ID = "summonerId"
    TIME = "time"

    AVAILABLE_CHAT_STATUSES = {CHAT := "chat", AWAY := "away", ONLINE := "online"}
