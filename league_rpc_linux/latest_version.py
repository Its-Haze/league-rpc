import requests

from league_rpc_linux.const import DDRAGON_API_VERSIONS


def get_latest_version() -> str:
    response = requests.get(url=DDRAGON_API_VERSIONS, timeout=15)

    data = response.json()
    latest_version = data[0]
    return latest_version
