import requests

from league_rpc.utils.const import DDRAGON_API_VERSIONS


def get_latest_version() -> str:
    try:
        response = requests.get(url=DDRAGON_API_VERSIONS, timeout=15)
        response.raise_for_status()
        data = response.json()
        latest_version = data[0]
        return latest_version
    except (requests.RequestException, ValueError, IndexError, KeyError) as e:
        raise RuntimeError(f"Failed to fetch latest version from DDragon API: {e}") from e
