import urllib3

from league_rpc_linux.polling import wait_until_exists
from league_rpc_linux.username import summoner_name

urllib3.disable_warnings()


def get_champion_name() -> str:
    """
    Get the current playing champion name.
    """
    champion_name = ""
    all_game_data_url = "https://127.0.0.1:2999/liveclientdata/allgamedata"
    your_summoner_name = summoner_name()

    if response := wait_until_exists(
        url=all_game_data_url,
        custom_message="Did not find the champion name. Will try again in 5 seconds",
    ):
        parsed_data = response.json()
        for player in parsed_data["allPlayers"]:
            if player["summonerName"] == your_summoner_name:
                champion_name = player["championName"]
                break
            continue

        print(f"Champion name found ({champion_name}), continuing...")

    match champion_name:
        case "Aurelion Sol":
            champion_name = "AurelionSol"
        case "Cho'Gath":
            champion_name = "Chogath"
        case "Renata Glasc":
            champion_name = "Renata"
        case "Dr. Mundo":
            champion_name = "DrMundo"
        case "Miss Fortune":
            champion_name = "MissFortune"
        case "Kai'Sa":
            champion_name = "KaiSa"
        case "Kog'Maw":
            champion_name = "KogMaw"
        case "Rek'Sai":
            champion_name = "RekSai"
        case "K'Sante":
            champion_name = "KSante"
        case "Kha'Zix":
            champion_name = "KhaZix"
        case "Nunu & Willump":
            champion_name = "Nunu"
        case "Twisted Fate":
            champion_name = "TwistedFate"
        case "Tahm Kench":
            champion_name = "TahmKench"
        case "Vel'Koz":
            champion_name = "Velkoz"
        case "Xin Zhao":
            champion_name = "XinZhao"
        case "Master Yi":
            champion_name = "MasterYi"
        case _:
            return champion_name

    return champion_name


def champion_asset(champion: str) -> str:
    """
    Returns the champion icon in .png format.
    """

    if (skin_id := get_champion_skin_id()) != 0:
        # Return the skin art
        print("Skin detected successfully!")
        return f"https://ddragon.leagueoflegends.com/cdn/img/champion/loading/{champion}_{skin_id}.jpg"

    # Otherwise return the default champion art.
    url = f"http://ddragon.leagueoflegends.com/cdn/13.10.1/img/champion/{champion}.png"
    print(f"Choosing Default skin on {champion}")
    return url


def get_champion_skin_id() -> int:
    """
    No skins = ID (0)
    Skins = ID (+1) -> max amount of skins on champ
    """
    url = "https://127.0.0.1:2999/liveclientdata/allgamedata"
    if response := wait_until_exists(
        url=url,
        custom_message="Could not find Game data. Will try again..",
    ):
        parsed_data = response.json()
        skin_id = parsed_data["allPlayers"][0]["skinID"]
        return skin_id

    return 0  # If no data of the skin_id was found.. return default skin.
