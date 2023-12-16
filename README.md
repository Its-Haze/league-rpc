
![Logo](https://github.com/daglaroglou/league-rpc-linux/blob/main/assets/league-rpc.png?raw=true)

![Linux](https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black) ![Python](https://img.shields.io/badge/Python-FCC624?style=for-the-badge&logo=python&logoColor=blue)

# League of Legends Discord Rich Presence for Linux (WINE / Lutris)

**Enhance your Discord experience while playing League of Legends on Linux!** This project brings unique Discord Rich Presence integration to Linux users, leveraging WINE/Lutris environments, with features not even available on Windows!

## Table of Contents

- [Installation](#installation)
  - [Source Code](#source-code)
  - [Package](#package)
- [Features](#features)
- [FAQ](#faq)
- [Contact and Support](#contact-and-support)
- [Credits](#credits)

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

## Features

1. **Dynamic Champion Skin Display**: The Discord Rich Presence will now show the champion skin you're using as the large image, a feature unique to this Linux version.
2. **Live KDA Stats**: Keep track of your Kills, Deaths, and Assists updated live during the game. (Future update will include an option to toggle this feature.)
3. **Creep Score**: Your minions (creep score) are displayed, providing a comprehensive view of your in-game performance.
4. **Precise In-Game Time Tracking**: The in-game time is calculated with precision. Even if the script stops, when restarted, it will display the correct in-game time, ensuring continuous and accurate representation of your game status.

## FAQ

**Q: Will this get my account banned?**
A: No, it uses Riot's local API at `127.0.0.1:2999`. You're responsible for its use.

**Q: Is this legal?**
A: This is an independent project, not affiliated or supported by Riot Games.

**Q: Does it support TFT?**
A: Currently, it does not support TFT.

**Q: Can it run on Windows?**
A: It might work with some adjustments, but Windows already has native support for this feature.

## Contact and Support

For any questions, feel free to contact me on Discord (@haze.dev) or open an issue on GitHub.

## Credits

Credit to the original repository where this project stems from: <https://github.com/daglaroglou/league-rpc-linux>

This is my version of league-rpc-linux, where I have added more functionality and also refactored most of the code, while also making it more `pythonic`.
