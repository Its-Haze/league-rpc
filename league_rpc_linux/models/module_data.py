from dataclasses import dataclass, field
from typing import Optional

from lcu_driver.connector import Connector
from pypresence import Presence

from league_rpc_linux.models.client_data import ClientData


# contains module internal data
@dataclass
class ModuleData:
    connector: Connector = field(default_factory=Connector)
    rpc: Optional[Presence] = None
    client_data: ClientData = field(default_factory=ClientData)
