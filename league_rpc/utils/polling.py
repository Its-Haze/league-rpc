import time

import requests
from urllib3.exceptions import NewConnectionError


def wait_until_exists(
    url: str,
    custom_message: str = "",
    expected_response_code: int = 200,
    timeout: int = 30,
    n_sleep: float | int = 5,  # Not needed, but good to have.
    n_total_amount: int = 20,
    startup: int = False,  # Set to True on the first time it tries to poll the local api. (onGameStart)
) -> requests.Response | None:
    """
    Polling on the local riot api until success is returned.
    """

    for _ in range(n_total_amount):
        try:
            response = requests.get(url, timeout=timeout, verify=False)
            if response.status_code != expected_response_code:
                time.sleep(n_sleep)
                continue
            break
        except (
            NewConnectionError,
            ConnectionError,
            requests.exceptions.ConnectionError,
        ):
            # These errors occur either before the api has started..
            # Or when the game has ended
            if startup:
                # Make sure we continue to poll the api during the start of a game.
                time.sleep(n_sleep)
                continue

            # When game ends, we don't care about polling the api.
            return None
    else:
        print(custom_message)
        return None
    return response
