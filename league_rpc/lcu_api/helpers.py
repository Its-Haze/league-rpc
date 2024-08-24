import time
from typing import Any

from lcu_driver.connection import Connection  # type:ignore


from league_rpc.champion import (
    gather_game_mode,
    gather_ingame_information,
    get_skin_asset,
)
from league_rpc.gametime import get_current_ingame_time
from league_rpc.kda import get_creepscore, get_kda, get_level
from league_rpc.lcu_api.base_data import gather_tft_companion_data_synchroneous

from league_rpc.models.module_data import ModuleData
from league_rpc.utils.const import (
    CHAMPION_NAME_CONVERT_MAP,
    LEAGUE_OF_LEGENDS_LOGO,
    SMALL_TEXT,
)
import rich
from league_rpc.models.client_data import ArenaStats, ClientData, RankedStats, TFTStats
import requests
from requests.auth import HTTPBasicAuth


async def get_current_state(connection: Connection) -> str:
    response = await connection.request(  # type:ignore
        "get", "/lol-gameflow/v1/gameflow-phase"
    )
    state = await response.json()
    return state


def get_current_state_sync(connection: Connection) -> str:
    """
    Fetch the current game flow state synchronously using requests.

    Args:
        connection (Connection): A connection object containing the address and authentication key.

    Returns:
        str: The current game flow state or an error message.
    """
    # Inspecting connection details for debugging

    # Define the endpoint
    endpoint = f"{connection.address}/lol-gameflow/v1/gameflow-phase"

    try:
        # Perform the synchronous GET request

        response = requests.get(
            endpoint,
            auth=HTTPBasicAuth("riot", connection.auth_key),
            verify=False,  # Disabling SSL verification; adjust as necessary
            timeout=15,  # Set a timeout to prevent hanging
        )

        # Raise an error if the response was not successful
        response.raise_for_status()

        # Parse the JSON response
        state = response.json()

        # Inspecting the response for debugging
        # rich.inspect(state)
        return state

    except requests.RequestException as e:
        rich.print(f"An error occurred: {e}")
        return "Error"


async def get_ingame_data(connection: Connection) -> dict[str, Any]:
    response = await connection.request(  # type:ignore
        "get", "/lol-gameflow/v1/session"
    )
    data = await response.json()
    return data


def show_ranked_data(
    module_data: ModuleData,
) -> tuple[str, ...]:
    """Helper method to fetch formatted ranked data for display in Rich Presence."""
    large_text = small_text = small_image = ""

    current_queue_name = module_data.client_data.get_queue_name

    match current_queue_name:
        case "Ranked Solo/Duo":
            summoner_rank: RankedStats = module_data.client_data.summoner_rank  # type: ignore
            if summoner_rank.tier:
                (
                    small_text,
                    small_image,
                ) = summoner_rank.rpc_info
                large_text = SMALL_TEXT

        case "Ranked Flex":
            summoner_rank = module_data.client_data.summoner_rank_flex  # type: ignore
            if summoner_rank.tier:
                (
                    small_text,
                    small_image,
                ) = summoner_rank.rpc_info
                large_text = SMALL_TEXT

        case "Teamfight Tactics (Ranked)":
            summoner_rank: TFTStats = module_data.client_data.tft_rank  # type: ignore
            if summoner_rank.tier:
                (
                    small_text,
                    small_image,
                ) = summoner_rank.rpc_info
                large_text = SMALL_TEXT
        case "Arena":
            summoner_rank: ArenaStats = module_data.client_data.arena_rank  # type: ignore
            if summoner_rank.tier:
                (
                    small_text,
                    small_image,
                ) = summoner_rank.rpc_info
                large_text = SMALL_TEXT

        case _:
            ...
    return large_text, small_image, small_text


def handle_in_game(
    connection: Connection,
    silent: bool,
    module_data: ModuleData,
) -> None:
    """
    Executes the appropriate function based on the current game mode.

    silent, is meant to not display console output if the inGame has already been ran

    """
    game_mode = gather_game_mode()

    if game_mode in (
        "Summoner's Rift (Custom)",
        "Summoner's Rift",
        "Summoner's Rift (Tutorial)",
        "Summoner's Rift (URF)",
        "Howling Abyss (ARAM)",
    ):
        handle_normal_game(silent, module_data)
    elif game_mode == "TFT":
        handle_tft_game(connection, silent, module_data)
    elif game_mode == "Arena":
        handle_arena_game(silent, module_data)
    elif game_mode == "Swarm - PVE":
        handle_swarm_game(silent, module_data)
    else:
        print(f"Unknown game mode: {game_mode}")
        return None


def handle_swarm_game(silent: bool, module_data: ModuleData) -> None:
    """
    Gather data specific to Swarm games
    """
    (
        champ_name,
        skin_name,
        chroma_name,
        skin_id,
        _,  # gamemode
        level,
        gold,
    ) = gather_ingame_information(silent=silent)
    skin_asset: str = get_skin_asset(
        champion_name=champ_name,
        skin_id=skin_id,
    )
    large_text = (
        f"{skin_name} ({chroma_name})"
        if chroma_name
        else (
            skin_name
            if skin_name
            else CHAMPION_NAME_CONVERT_MAP.get(champ_name, champ_name)
        )
    )
    module_data.rpc.update(  # type:ignore
        large_image=skin_asset,
        large_text=large_text,
        details=module_data.client_data.get_queue_name,
        state=f"In Game {f'· {get_creepscore()} · lvl: {level} · gold: {gold}' if not module_data.cli_args.no_stats else ''}",
        small_image=LEAGUE_OF_LEGENDS_LOGO,
        small_text=SMALL_TEXT,
        start=int(time.time())
        - get_current_ingame_time(default_time=module_data.start_time),
    )


def handle_arena_game(silent: bool, module_data: ModuleData) -> None:
    """
    Gather data specific to Arena games
    """
    (
        champ_name,
        skin_name,
        chroma_name,
        skin_id,
        _,  # gamemode
        level,
        gold,
    ) = gather_ingame_information(silent=silent)

    skin_asset: str = get_skin_asset(
        champion_name=champ_name,
        skin_id=skin_id,
    )
    large_text = (
        f"{skin_name} ({chroma_name})"
        if chroma_name
        else (
            skin_name
            if skin_name
            else CHAMPION_NAME_CONVERT_MAP.get(champ_name, champ_name)
        )
    )

    module_data.rpc.update(  # type:ignore
        large_image=skin_asset,
        large_text=large_text,
        details=module_data.client_data.get_queue_name,
        state=f"In Game {f'· {get_kda()} · lvl: {level} · gold: {gold}' if not module_data.cli_args.no_stats else ''}",
        small_image=LEAGUE_OF_LEGENDS_LOGO,
        small_text=SMALL_TEXT,
        start=int(time.time())
        - get_current_ingame_time(default_time=module_data.start_time),
    )


def handle_normal_game(silent: bool, module_data: ModuleData) -> None:
    """
    Gather data specific to summoners rift games
    """
    (
        champ_name,
        skin_name,
        chroma_name,
        skin_id,
        gamemode,
        _,
        _,
    ) = gather_ingame_information(silent=silent)

    skin_asset = get_skin_asset(
        champion_name=champ_name,
        skin_id=skin_id,
    )
    if not champ_name or not gamemode:
        return
    large_text = (
        f"{skin_name} ({chroma_name})"
        if chroma_name
        else (
            skin_name
            if skin_name
            else CHAMPION_NAME_CONVERT_MAP.get(champ_name, champ_name)
        )
    )
    small_image = LEAGUE_OF_LEGENDS_LOGO
    small_text = SMALL_TEXT
    module_data.client_data.queue_detailed_description = "Ranked Flex"
    if not module_data.cli_args.no_rank:  # type: ignore
        _, _small_image, _small_text = show_ranked_data(module_data)
        if all([_small_image, _small_text]):
            small_image, small_text = (
                _small_image,
                _small_text,
            )

    module_data.rpc.update(  # type:ignore
        large_image=skin_asset,
        large_text=large_text,
        details=module_data.client_data.get_queue_name,
        state=f"In Game {f'· {get_kda()} · {get_creepscore()}' if not module_data.cli_args.no_stats else ''}",
        small_image=small_image,
        small_text=small_text,
        start=int(time.time())
        - get_current_ingame_time(default_time=module_data.start_time),
    )


def handle_tft_game(
    connection: Connection, silent: bool, module_data: ModuleData
) -> None:
    if not silent:
        # only gather this data once.
        gather_tft_companion_data_synchroneous(connection, module_data.client_data)

    module_data.rpc.update(  # type:ignore
        large_image=module_data.client_data.tft_companion_icon,
        large_text=module_data.client_data.tft_companion_name,
        details=module_data.client_data.get_queue_name,
        state=f"In Game · lvl: {get_level()}",
        small_image=LEAGUE_OF_LEGENDS_LOGO,
        small_text=SMALL_TEXT,
        start=int(time.time())
        - get_current_ingame_time(default_time=module_data.start_time),
    )
