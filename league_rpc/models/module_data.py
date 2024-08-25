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

from argparse import Namespace
from dataclasses import dataclass, field
import time
from typing import Optional

from lcu_driver.connector import Connector
from pypresence import Presence

from league_rpc.logger.richlogger import RichLogger
from league_rpc.models.client_data import ClientData


# contains module internal data
@dataclass
class ModuleData:
    """A dataclass designed to store the operational state of a module, including connections to the
    League client and the current state of any ongoing Rich Presence integrations.
    """

    connector: Connector = field(default_factory=Connector)
    client_data: ClientData = field(default_factory=ClientData)
    logger: RichLogger = field(default_factory=RichLogger)
    rpc: Optional[Presence] = None
    cli_args: Optional[Namespace] = None
    start_time = int(time.time())
