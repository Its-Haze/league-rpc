
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

2.

   ```
   ./league_rpc_linux
   ```

3. Yup.. that's pretty much it! Enjoy

## Features

1. **Dynamic Champion Skin Display**: The Discord Rich Presence will now show the champion skin you're using as the large image, a feature unique to this Linux version.
2. **Live KDA Stats**: Keep track of your Kills, Deaths, and Assists updated live during the game. (Future update will include an option to toggle this feature.)
3. **Creep Score**: Your minions (creep score) are displayed, providing a comprehensive view of your in-game performance.
4. **Precise In-Game Time Tracking**: The in-game time is calculated with precision. Even if the script stops, when restarted, it will display the correct in-game time, ensuring continuous and accurate representation of your game status.

## FAQ

**Q: Will this get my account banned?**
A: No, it uses Riot's local API at `127.0.0.1:2999`. You're responsible for its use.

**Q: Is this legal?**
A: This is an independent project, not affiliated with Riot Games. If they like the concept of showing Skins, KDA+cs they are more than welcome to steal my ideas.

**Q: Does it support TFT?**
A: Yes! TFT is supported as of v0.0.2. Just start a TFT game, and the script will automagically detect it and show that you are ingame, with a neat "lvl" indicator.

**Q: Can it run on Windows?**
A: It might work with some adjustments, but Windows already has native support for this feature. Although their discord rich presense is not as beautiful as this one.. seriously!

## Contact and Support

For any questions, feel free to contact me on Discord (@haze.dev) or open an issue on GitHub.

## Credits

Credit to the original repository where this project stems from: <https://github.com/daglaroglou/league-rpc-linux>

This is my version of league-rpc-linux, where I have added more functionality and also refactored the code to fit best practices and pep standards.
