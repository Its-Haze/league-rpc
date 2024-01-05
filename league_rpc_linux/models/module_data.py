from argparse import Namespace
from dataclasses import dataclass, field
from typing import Optional

from lcu_driver.connector import Connector
from pypresence import Presence

from league_rpc_linux.models.client_data import ClientData


# contains module internal data
@dataclass
class ModuleData:
    connector: Connector = field(default_factory=Connector)
    client_data: ClientData = field(default_factory=ClientData)
    rpc: Optional[Presence] = None
    cli_args: Optional[Namespace] = None
