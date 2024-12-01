from dataclasses import dataclass


@dataclass
class RPCData:
    """A dataclass that stores the data required to update the Discord Rich Presence."""

    large_image: str = ""
    large_text: str = ""
    small_image: str = ""
    small_text: str = ""
    details: str = ""
    state: str = ""
    start: int = 0
