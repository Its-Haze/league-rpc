"""
This module defines classes for encapsulating various stages of the game flow and lobby/player statuses
in League of Legends, interfacing with two specific API endpoints: /lol-summoner/v1/current-summoner
for game flow phases and /lol-gameflow/v1/gameflow-metadata/player-status for lobby and player statuses.
These classes organize data relevant to game progression and lobby management, ensuring it can be accessed
and manipulated efficiently.

Usage:
    These classes are crucial for developing features that require tracking or responding to changes in the game's
    status or managing interactions within the game lobby. They facilitate the creation of responsive and interactive
    features in client tools, enhancing the user experience by providing real-time updates and controls based on the
    game flow and lobby conditions. This structured approach helps maintain clarity and consistency when interacting
    with the related APIs, supporting a robust development environment for League of Legends client applications.
"""


class GameFlowPhase:
    """Enumerates all possible phases of the game flow, from lobby formation to the end of a game,
    providing clear constants to represent each phase of gameplay.
    """

    NONE = "None"
    LOBBY = "Lobby"
    MATCHMAKING = "Matchmaking"
    CHECKED_INTO_TOURNAMENT = "CheckedIntoTournament"
    READY_CHECK = "ReadyCheck"
    CHAMP_SELECT = "ChampSelect"
    GAME_START = "GameStart"
    FAILED_TO_LAUNCH = "FailedToLaunch"
    IN_PROGRESS = "InProgress"
    RECONNECT = "Reconnect"
    WAITING_FOR_STATS = "WaitingForStats"
    PRE_END_OF_GAME = "PreEndOfGame"
    END_OF_GAME = "EndOfGame"
    TERMINATED_IN_ERROR = "TerminatedInError"


class LolGameflowLobbyStatus:
    """Contains attributes that describe the status of a game lobby, including whether
    it's a custom game, the lobby's queue ID, and details about its members and spectators.
    """

    ALLOWED_PLAY_AGAIN = "allowedPlayAgain"
    CUSTOM_SPECTATOR_POLICY = "customSpectatorPolicy"
    INVITED_SUMMONER_IDS = "invitedSummonerIds"
    IS_CUSTOM = "isCustom"
    IS_LEADER = "isLeader"
    IS_PRACTICE_TOOL = "isPracticeTool"
    IS_SPECTATOR = "isSpectator"
    LOBBY_ID = "lobbyId"
    MEMBER_SUMMONER_IDS = "memberSummonerIds"
    QUEUE_ID = "queueId"


class LolGameflowPlayerStatus:
    """Holds fields related to the player's status within the game flow, detailing their
    ability to invite others post-game and status of their current and last queued lobbies.
    """

    CAN_INVITE_OTHERS_AT_EOG = "canInviteOthersAtEog"
    CURRENT_LOBBY_STATUS = "currentLobbyStatus"
    LAST_QUEUED_LOBBY_STATUS = "lastQueuedLobbyStatus"
