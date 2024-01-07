"""
This module contains the fields for extracting information from

GameFlowPhase: /lol-summoner/v1/current-summoner

LolGameflowLobbyStatus: /lol-gameflow/v1/gameflow-metadata/player-status
LolGameflowPlayerStatus: /lol-gameflow/v1/gameflow-metadata/player-status
"""


class GameFlowPhase:
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
    CAN_INVITE_OTHERS_AT_EOG = "canInviteOthersAtEog"
    CURRENT_LOBBY_STATUS = "currentLobbyStatus"
    LAST_QUEUED_LOBBY_STATUS = "lastQueuedLobbyStatus"
