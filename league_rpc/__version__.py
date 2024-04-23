"""Set version number of package."""

__version__ = "v2.0.0"
import requests

RELEASES_PAGE = "https://github.com/Its-Haze/league-rpc/releases"


def get_version_from_github() -> str:
    """
    Get the latest version of the software from the GitHub repository.

    Returns:
        str: The latest version of the software.
    """
    response: requests.Response = requests.get(
        url="https://api.github.com/repos/its-haze/league-rpc/releases/latest",
        timeout=15,
    )
    return response.json()["tag_name"]


def check_latest_version() -> bool:
    """
    Check if the current version of the software is the latest version available.

    Returns:
        bool: True if the current version is not the latest version, False otherwise.
    """
    latest_version: str = get_version_from_github()
    return latest_version > __version__
