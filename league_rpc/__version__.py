"""Set version number of package."""

__version__ = "v3.2.0"
import requests

RELEASES_PAGE = "https://github.com/Its-Haze/league-rpc/releases"


def get_version_from_github() -> str | None:
    """
    Get the latest version of the software from the GitHub repository.

    Returns:
        str: The latest version of the software.
    """
    response: requests.Response = requests.get(
        url="https://api.github.com/repos/its-haze/league-rpc/releases/latest",
        timeout=15,
    )
    if response.status_code != 200:
        # Probably due to rate limit.. should be handled better.
        return None

    return response.json()["tag_name"]


def check_latest_version(latest_version: str) -> bool | None:
    """
    Check if the current version of the software is the latest version available.

    Returns:
        bool: True if the current version is not the latest version, False otherwise.
    """
    return latest_version > __version__
