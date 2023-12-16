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

    try:
        for _ in range(n_total_amount):
            response = requests.get(url, timeout=timeout, verify=False)
            if response.status_code != expected_response_code:
                time.sleep(n_sleep)
                continue
            break
        else:
            print(custom_message)
            return None
    except (NewConnectionError, ConnectionError, requests.exceptions.ConnectionError):
        # Ignore these.. These will come when you finish games.
        return None

    return response
