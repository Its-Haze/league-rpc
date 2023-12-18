
![Logo](https://github.com/its-haze/league-rpc-linux/blob/master/assets/league-rpc.png?raw=true)

![Linux](https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black) ![Python](https://img.shields.io/badge/Python-FCC624?style=for-the-badge&logo=python&logoColor=blue)

# League of Legends Discord Rich Presence for Linux (WINE / Lutris)

**Enhance your Discord experience while playing League of Legends on Linux!** This project brings unique Discord Rich Presence integration to Linux users, leveraging WINE/Lutris environments, with features not even available on Windows!

## Table of Contents

- [Showcase](#showcase)
- [Installation](#installation)
  - [Source Code](#source-code)
  - [Package](#package)
- [Features](#features)
- [Command Line Arguments](#command-line-arguments)
- [FAQ](#faq)
- [Contact and Support](#contact-and-support)
- [Credits](#credits)

## Showcase

### League of legends

#### Default skin example

![Garen-Default](https://github.com/Its-Haze/league-rpc-linux/blob/master/assets/league-rpc-example-3.png?raw=true)

#### Skin example

Just pick a skin in champ select and the script will automagically detect it and update discord RPC when your game starts!

![Garen-Skin](https://github.com/Its-Haze/league-rpc-linux/blob/master/assets/league-rpc-example-2.png?raw=true)

#### KDA + Minnions calculated and presented in Discord

![MissFortune-KDA-CS](https://github.com/Its-Haze/league-rpc-linux/blob/master/assets/league-rpc-skin-example.png?raw=true)

### TFT (Teamfight tactics)

TFT is supported and will show a cuter image than windows ever will! :)

![Teamfight-Tactics-example](https://github.com/Its-Haze/league-rpc-linux/blob/master/assets/tft-rpc-example.jpeg?raw=true)

## Installation

### Source Code

1. Clone the repository:

   ```bash
   git clone git@github.com:Its-Haze/league-rpc-linux.git
   ```

2. Install dependencies:

   ```bash
   pip3 install -r requirements.txt
   ```

3. Run the application:

   ```bash
   python3 -m league_rpc_linux
   ```

### Package

Alternatively, install the executable Python package directly from the [releases page](https://github.com/Its-Haze/league-rpc-linux/releases) and run it.

1. Download the executable file found in [releases page](https://github.com/Its-Haze/league-rpc-linux/releases)

   ```bash
   wget https://github.com/Its-Haze/league-rpc-linux/releases/download/v0.0.5/league_rpc_linux
   ```

2. Make the file executable

   ```bash
   chmod +x ./league_rpc_linux
   ```

3. Run it

   ```bash
   ./league_rpc_linux
   ```

4. Yup.. that's pretty much it! Enjoy

## Command Line Arguments

This application now supports various command-line arguments to enhance flexibility and user customization.

- **`--client-id`**: Specify a custom Discord client ID for the RPC connection. Defaults to `1185274747836174377` if not provided. Which is the "League of Linux" Application
  - *Example*: `./league_rpc_linux --client-id 123456789123456789`

- **`--no-stats`**: Opt out of displaying in-game KDA and minion (creep score) statistics in your Discord Rich Presence. By default, these stats are shown.
  - *Example*: `./league_rpc_linux --no-stats`

- **`--add-process`**: Add custom Discord process names to the search list. This is useful if your Discord client is running under a different process name.
  - *Example*: `./league_rpc_linux --add-process CustomDiscord AnotherProcess`

Each of these arguments can be combined to tailor the Discord RPC to your preferences.

```bash
./league_rpc_linux --client-id 123456789123456789 --no-stats --add-process CustomDiscord AnotherProcess
```

NOTE: All of these arguments are optional. None is needed for the script to function properly. But in case you want to change something, you now can.

## Features

1. **Dynamic Champion Skin Display**: The Discord Rich Presence will now show the champion skin you're using as the large image, a feature unique to this Linux version.
2. **Live KDA Stats**: Keep track of your Kills, Deaths, and Assists updated live during the game. (Future update will include an option to toggle this feature.)
3. **Creep Score**: Your minions (creep score) are displayed, providing a comprehensive view of your in-game performance.
4. **Precise In-Game Time Tracking**: The in-game time is calculated with precision. Even if the script stops, when restarted, it will display the correct in-game time, ensuring continuous and accurate representation of your game status.
5. **Discord Reconnection Window**: While the script only works while Discord is up and running. There are instances where discord could crash.. This program will attempt to reconnect to Discord's RPC even if the app is not running. For a period of time, and only exit when too much time has passed.. Default is (50) seconds to restart/reconnect Discord.

## FAQ

**Q: Will this get my account banned?**
A: No, it uses Riot's local API at `127.0.0.1:2999`. You're responsible for its use.

**Q: Is this legal?**
A: This is an independent project, not affiliated with Riot Games. If they like the concept of showing Skins, KDA+cs they are more than welcome to steal my ideas.

**Q: Does it support TFT?**
A: Yes! TFT is supported as of v0.0.2. Just start a TFT game, and the script will automagically detect it and show that you are ingame, with a neat "lvl" indicator.

**Q: Can it run on Windows?**
A: It might work with some adjustments, but Windows already has native support for this feature. Although their discord rich presense is not as beautiful as this one.. seriously!

**Q: Why doesn't the minnions (cs) update live?**
A: Trust me it's annoying for me as well.. This is thanks to Riot Games not updating their API for every minnion killed, but rather updates every (10th) minnion killed. So this is sadly out of my control.

## Contact and Support

For any questions, feel free to contact me on Discord (@haze.dev) or open an issue on GitHub.

## Credits

Credit to the original repository where this project stems from: <https://github.com/daglaroglou/league-rpc-linux>

This is my version of league-rpc-linux, where I have added more functionality and also refactored the code to fit best practices and pep standards.
