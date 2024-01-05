"""
This module contains the fields for extracting information from /lol-chat/v1/me
"""


class LolChatUser:
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
