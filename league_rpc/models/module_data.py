"""
This module defines the ModuleData class, which holds essential internal state data and connections necessary
for interacting with the League of Legends client via the LCU (League Client Update) Driver. This class facilitates
the integration of client data into external applications, particularly those that enhance in-game interactions or
functionality through additional overlays or tools.

Usage:
    The ModuleData class is integral to applications that interact with the League of Legends client, providing
    a centralized repository for managing connections and state. It is especially useful in environments where
    multiple components or services must access or modify the client state or where integration with third-party
    services like Discord for Rich Presence is required. This setup supports a robust, maintainable codebase
    by ensuring that essential state and connection information is easily accessible and systematically organized.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Optional

if TYPE_CHECKING:
    from league_rpc.models.rpc_data import RPCData
    from league_rpc.models.client_data import ClientData
    from league_rpc.models.rpc_updater import RPCUpdater

import time
from argparse import Namespace
from dataclasses import dataclass, field

from lcu_driver.connector import Connector
from pypresence import Presence

from league_rpc.logger.richlogger import RichLogger


# contains module internal data
@dataclass
class ModuleData:
    """A dataclass designed to store the operational state of a module, including connections to the
    League client and the current state of any ongoing Rich Presence integrations.
    """

    client_data: "ClientData"

    # Triggering/scheduling the update of the Discord Rich Presence
    rpc_updater: "RPCUpdater"

    # Storing data regarding the Discord Rich Presence currently active
    rpc_data: "RPCData"

    connector: Connector = field(default_factory=Connector)
    logger: RichLogger = field(default_factory=RichLogger)
    cli_args: Optional[Namespace] = None
    start_time = int(time.time())

    # Discord Rich Presence instance from pypresence
    rpc: Optional[Presence] = None
