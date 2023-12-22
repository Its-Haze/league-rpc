import requests
from league_rpc_linux.processes.process import get_process_command


class LCU:
    """
    League Client Update (LCU) API class.
    """

    def __init__(self):
        self.lcu_connect = None
        self.lcu_connect_port = None
        self.lcu_connect_password = None
        self.lcu_connect_protocol = "https"
        self.lcu_connect_host = "127.0.0.1"
        self.get_lcu_connect_details(proc_name="CrBrowserMain")

    def get_lcu_connect_details(self, proc_name: str) -> bool:
        """
        Get LCU connect details from command.
        """

        cmd = get_process_command(proc_name)

        if not cmd:
            return False

        for arg in cmd:
            if "--remoting-auth-token=" in arg:
                self.lcu_connect_password = arg.split("=")[1]
            elif "--app-port=" in arg:
                self.lcu_connect_port = arg.split("=")[1]
        return True

    def reloadConnectDetails(self):
        """
        Reload the connect details.
        """
        return self.get_lcu_connect_details(
            proc_name="CrBrowserMain"
        ) | self.get_lcu_connect_details(proc_name="LeagueClientUx.exe")

    def getDataFromPath(self, path: str) -> dict:
        """
        Get data from path.
        """
        url = f"https://riot:{self.lcu_connect_password}@{self.lcu_connect_host}:{self.lcu_connect_port}{path}"

        response = requests.get(url=url, verify=False)

        if response.status_code == 200:
            return response.json()

        # Auto reload connect details if 403 - maybe client got closed during game?
        if response.status_code == 403:
            if self.reloadConnectDetails():
                return self.getDataFromPath(path)
            else:
                print(
                    "Failed to reload connect details. - Is your league client running?"
                )

        return False


lcuApi = None


def getLcuAPI():
    """
    Get LCU API.
    """
    global lcuApi
    if not lcuApi:
        lcuApi = LCU()
    return lcuApi
