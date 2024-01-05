
![Logo](https://github.com/its-haze/league-rpc-linux/blob/master/assets/league-rpc.png?raw=true)

![Linux](https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black)
![Windows](https://img.shields.io/badge/Windows-FCC624?style=for-the-badge&logo=Windows&logoColor=black)
![Mac](https://img.shields.io/badge/Mac-FCC624?style=for-the-badge&logo=Apple&logoColor=black)
![Python](https://img.shields.io/badge/Python-FCC624?style=for-the-badge&logo=python&logoColor=blue)

# League of Legends Discord Rich Presence for Linux (WINE / Lutris)

**Enhance your Discord experience while playing League of Legends on Linux!** This project brings unique Discord Rich Presence integration to Linux users, leveraging WINE/Lutris environments, with features not even available natively on Windows!

## Table of Contents

- [Showcase](#showcase)
- [Installation](#installation)
  - [Source Code](#source-code)
  - [Pre-built Package](#pre-built-package)
  - [League RPC Auto Launcher Script (Lutris example)](#league-rpc-auto-launcher-script)
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

### Arena

Arena games are now supported as of v0.0.7

![Arena-example](https://github.com/Its-Haze/league-rpc-linux/blob/master/assets/league-rpc-arena-example.png?raw=true)

## Installation

- Are you using Lutris? Then please take a look at [League RPC Auto Launcher Script (Lutris example)](#league-rpc-auto-launcher-script)

### Source Code

1. **Clone the repository**:

   Use the following command to clone the repository to your local machine:

   ```bash
   git clone git@github.com:Its-Haze/league-rpc-linux.git
   ```

2. **Create and activate a virtual environment**:

   A virtual environment helps to keep dependencies required by different projects separate by creating isolated python environments for them. This is highly recommended to avoid conflicts between project dependencies.

   - Create a virtual environment:

     ```bash
     python3 -m venv league-rpc-env
     ```

   - Activate the virtual environment:
     - On Unix or MacOS:

       ```bash
       source league-rpc-env/bin/activate
       ```

     - On Windows:

       ```bash
       .\league-rpc-env\Scripts\activate
       ```

   For more information on virtual environments and why they are beneficial, you can read this [Python Virtual Environments Guide](https://docs.python.org/3/tutorial/venv.html).

3. **Install dependencies**:

   With the virtual environment activated, install the required dependencies using:

   ```bash
   pip install -r requirements.txt
   ```

4. **Run the application**:

   Finally, start the application with:

   ```bash
   python3 -m league_rpc_linux
   ```

### Pre-built Package

Alternatively, install the executable Python package directly from the [releases page](https://github.com/Its-Haze/league-rpc-linux/releases) and run it.

1. Go to the [releases page](https://github.com/Its-Haze/league-rpc-linux/releases) page and download the latest executable file ``league_rpc_linux``
   - These builds are mainly built on Unix systems.. meaning that they will **not** work on Windows. If you see a .exe file, then you are good to download that one instead.

2. Once downloaded, make the file executable. Open a terminal and run the following command (assuming the file was downloaded to the current directory):

   ```bash
   chmod +x ./league_rpc_linux
   ```

3. Run it

   ```bash
   ./league_rpc_linux
   ```

4. Yup.. that's pretty much it! Enjoy

### League RPC Auto Launcher Script

The `league-rpc-auto-launcher.sh` script enhances your League of Legends experience on Linux by automatically updating and launching the League of Legends Discord Rich Presence alongside the game.

This script can be ran as a standalone, but it is recommended to be put as a Pre-launch script for launchers such as Lutris.

Before running any bash script from strangers on the internet.. Please read and try to understand what it does.
Short summary:

- It will download the latest version of `league_rpc_linux`
- It will then execute this program.
- If you don't want it to always fetch the latest version of `league_rpc_linux`, then change the variable `AUTO_INSTALL` to `false` like this `AUTO_INSTALL=false` (default is true).
- If you have set the variable `$XDG_BIN_HOME` then that will be used to store the executable. otherwise `$HOME/.local/bin` will be used.
- If you have set the variable `$XDG_CACHE_HOME` then that will be used to store the log file. otherwise `$HOME/.cache/` will be used. Full path to the logfile is `$XDG_CACHE_HOME/league-rpc-linux/update_log.txt`

Follow these steps to set it up:

1. **Download the Script**:

   Download the `league-rpc-auto-launcher.sh` script

   ```bash
   wget https://raw.githubusercontent.com/Its-Haze/league-rpc-linux/master/league-rpc-auto-launcher.sh -O league-rpc-auto-launcher.sh
   ```

2. **Make the Script Executable**:

   Once downloaded, you need to make the script executable. Open a terminal in the directory where the script is and run:

   ```bash
   chmod +x league-rpc-auto-launcher.sh
   ```

3. **Setting up in Lutris**:

   To have the script run automatically when you start League of Legends via Lutris:

   - Open Lutris and right-click on your League of Legends game.
   - Select '`Configure`' from the context menu.
   - In the configuration window, navigate to the '`System options`' tab.
   - Make sure you have '`Advanced`' mode enabled.
   - Scroll down to the '`Pre-launch script`' field.
   - Enter the full path to the `league-rpc-auto-launcher.sh` script. For example:

     ```
     /home/yourusername/path/to/league-rpc-auto-launcher.sh
     ```

   - Click '`Save`' to apply the changes.

Now, every time you start League of Legends through Lutris, the `league-rpc-auto-launcher.sh` script will ensure that the latest version of the Discord Rich Presence tool is installed and running, enhancing your gaming experience.

## Command Line Arguments

This application now supports various command-line arguments to enhance flexibility and user customization.

- **`--client-id`**: Specify a custom Discord client ID for the RPC connection. Defaults to `1185274747836174377` if not provided. Which is the "League of Linux" Application
  - *Example*: `./league_rpc_linux --client-id 123456789123456789`

- **`--no-stats`**: Opt out of displaying in-game KDA and minion (creep score) statistics in your Discord Rich Presence. By default, these stats are shown.
  - *Example*: `./league_rpc_linux --no-stats`

- **`--show-rank`**: Show off your League rank on Discord (SoloQ/Flex/TFT/Arena) By default, this will be hidden.
  - *Example*: `./league_rpc_linux --show-rank`

- **`--show-emojis`**: Do you want to show your Online/Away status with a emoji, then add this argument. By default, this will be hidden.
  - *Example*: `./league_rpc_linux --show-emojis`

- **`--add-process`**: Add custom Discord process names to the search list. This is useful if your Discord client is running under a different process name.
  - *Example*: `./league_rpc_linux --add-process CustomDiscord AnotherProcess`

- **`--wait-for-league`**: Specify the time (in seconds) the script should wait for the League of Legends client to start. Use `-1` for infinite waiting. This is particularly useful for auto-launch scenarios like with Lutris or other launchers, ensuring the script does not error out if League is not immediately detected.
  - *Example*: `./league_rpc_linux --wait-for-league 30`

- **`--wait-for-discord`**: Similar to `--wait-for-league`, specify the time (in seconds) to wait for Discord to start. Use `-1` for infinite waiting. This ensures that the script waits for Discord to fully start, avoiding premature errors.
  - *Example*: `./league_rpc_linux --wait-for-discord 15`

Each of these arguments can be combined to tailor the Discord RPC to your preferences.

```bash
./league_rpc_linux --client-id 123456789123456789 --no-stats --add-process CustomDiscord --wait-for-league -1 --wait-for-discord 15 --show-emojis --show-rank
```

Recommended flags:

```bash
./league_rpc_linux --wait-for-league -1 --wait-for-discord -1 --show-emojis --show-rank
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
A: Yes, it can. But Windows already has native support for league. Although their discord rich presense is not as beautiful as this one.. seriously! Even if you're on Windows, I recommend you try this out.

**Q: Why doesn't the minnions (cs) update live?**
A: Trust me it's annoying for me as well.. This is thanks to Riot Games not updating their API for every minnion killed, but rather updates every (10th) minnion killed. So this is sadly out of my control.

## Contact and Support

For any questions, feel free to contact me on Discord (@haze.dev) or open an issue on GitHub.

## Credits

Credit to the original repository where this project stems from: <https://github.com/daglaroglou/league-rpc-linux>

This is my version of league-rpc-linux, where I have added more functionality and also refactored the code to fit best practices and pep standards.
