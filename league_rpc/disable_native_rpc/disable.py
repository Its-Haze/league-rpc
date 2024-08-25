import json
import os
from typing import Any, Optional

import psutil

from league_rpc.logger.richlogger import RichLogger
from league_rpc.utils.color import Color

LEAGUE_NATIVE_RPC_PLUGIN = "rcp-be-lol-discord-rp"


def check_plugin_status(
    file_path: str,
    logger: RichLogger,
    plugin_name: str = LEAGUE_NATIVE_RPC_PLUGIN,
) -> None:
    """Check if a specific plugin is still in the manifest file and inform the user."""
    data: dict[str, Any] | None = load_json_file(file_path=file_path)
    if data is None:
        return

    plugin_found: bool = any(
        plugin["name"] == plugin_name for plugin in data.get("plugins", [])
    )
    if plugin_found:
        logger.warning(
            "The Native League Presence is still active. Please start this application before launching League of legends to fully disable it.",
            color="yellow",
        )



def load_json_file(file_path: str) -> Optional[dict[str, Any]]:
    """Load a JSON file from the provided path."""
    try:
        with open(file=file_path, mode="r+", encoding="utf-8") as file:
            return json.load(fp=file)
    except FileNotFoundError:
        print(f"{Color.red}No JSON file at {file_path}{Color.reset}")
        return None


def save_json_file(file_path: str, data: dict[str, Any]) -> None:
    """Save a dictionary to a JSON file."""
    with open(file=file_path, mode="w+", encoding="utf-8") as file:
        json.dump(obj=data, fp=file, indent=4)
        file.truncate()


def modify_json_data(
    data: dict[str, Any],
    plugin_name: str = LEAGUE_NATIVE_RPC_PLUGIN,
) -> bool:
    """Remove the specified plugin from the data."""
    modified = False
    for plugin in data.get("plugins", []):
        if plugin["name"] == plugin_name:
            data["plugins"].remove(plugin)
            modified = True
    return modified


def check_and_modify_json(file_path: str, logger: RichLogger) -> None:
    """Remove a specific plugin from the League manifest file."""
    data: dict[str, Any] | None = load_json_file(file_path=file_path)
    if data is None:
        return

    if modify_json_data(data=data):
        save_json_file(file_path=file_path, data=data)
        logger.info("Successfully disabled League Native Rich Presence")
        logger.update_progress_bar(advance=10)



def find_game_locale(league_processes: list[str]) -> str:
    """Find the locale, en_US, or something else of the current league process."""

    for proc in psutil.process_iter(attrs=["cmdline", "name"]):
        try:
            if proc.info["name"] in league_processes:
                locale_str: str = [
                    x for x in proc.info["cmdline"] if x.startswith("--locale=")
                ][0]
                locale: str = locale_str.split("=")[1]
                return locale
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            continue

    return "en_US"


def find_game_path() -> Optional[str]:
    """Find the path to the plugin-manifest.json file for League of Legends."""
    target_process = "RiotClientServices.exe"
    riot_path_identifier = "Riot Games"

    for proc in psutil.process_iter(attrs=["pid", "name", "exe"]):
        try:
            if (
                proc.info["name"] == target_process
                and riot_path_identifier in proc.info["exe"]
            ):
                base_path: Any = (
                    proc.info["exe"].split(riot_path_identifier)[0]
                    + riot_path_identifier
                )
                return os.path.join(
                    base_path, "League of Legends", "Plugins", "plugin-manifest.json"
                )
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            continue
    return None
