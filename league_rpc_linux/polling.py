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
) -> requests.Response | None:
    """
    Polling on the local riot api until success is returned.
    """

    max_failed = 0

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
            # Ignore these errors.. they occur when you exit a game.
            max_failed += 1
            if max_failed >= 3:
                return None
            continue
    else:
        print(custom_message)
        return None
    return response
