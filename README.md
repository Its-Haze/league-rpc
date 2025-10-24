<img src="https://github.com/its-haze/league-rpc/blob/master/assets/league-rpc.png?raw=true" width=30% height=auto />

<div align="left">

<a href="https://github.com/its-haze/league-rpc/releases/latest">![GitHub All Releases](https://img.shields.io/github/downloads/its-haze/league-rpc/total.svg?style=for-the-badge&color=purple)</a> <a href="https://github.com/its-haze/league-rpc/stargazers">![Stargazers](https://img.shields.io/github/stars/its-haze/league-rpc.svg?style=for-the-badge)</a> <a href="https://github.com/its-haze/league-rpc/issues">![Issues](https://img.shields.io/github/issues/its-haze/league-rpc.svg?style=for-the-badge)</a> <a href="https://github.com/Its-Haze/league-rpc/blob/master/LICENSE.txt">![MIT License](https://img.shields.io/github/license/its-haze/league-rpc.svg?style=for-the-badge)</a>

</div>
⭐ Don't forget to star this project! ⭐

# A Better League of Legends Rich Presence For Discord!

**Enhance your Discord experience while playing League of Legends!** This project brings unique Discord Rich Presence integration for League players, including features not even available natively from the game!


Got questions already? Don't hesitate to join the [Discord Community Server](https://discord.haze.sh)

## Table of Contents
- [Installation](#installation)
- [Showcase](#showcase)
- [Command Line Arguments](#command-line-arguments)
- [Tips for Running](#-tips-for-running)
- [FAQ](#faq)
- [Build from source](#️-build-from-source)
- [Contact and Support](#-contact-and-support)
- [History](#-history)

---

## Installation

### ✅ Prerequisites
I recommend using [Windows Terminal](https://aka.ms/terminal) and setting it as your [default terminal](https://devblogs.microsoft.com/commandline/windows-terminal-as-your-default-command-line-experience/). It makes everything look way better with colors and proper formatting instead of the boring old `cmd` window.

### 📥 Getting Started
1. Head over to the [Releases Page](https://github.com/Its-Haze/league-rpc/releases)
2. Download `leagueRPC.exe` from the latest release (it's under Assets)
3. Run the exe file (probably sitting in your Downloads folder)
4. Accept the Windows security popup if it shows up
5. That's it! ✨

### 🔄 Updating
No automatic updates yet, so you'll need to manually download new versions from the [Releases Page](https://github.com/Its-Haze/league-rpc/releases). Don't worry though - LeagueRPC will let you know in the terminal when there's a newer version available. I recommend staying up to date for the best experience.

---

## Showcase

### Summoner Icons

Who let the Kitten and the Penguin out? I did 😎. Now you too, can show off your favorite summoner icon, right there on Discord!

![summoner-icon-1](images/in_client_icon_1.png) ![summoner-icon-2](images/in_client_icon_2.png)

### Ranked Games

You can show off your rank emblem right in your Discord Presence.
- SoloQ/Flex: Shows off your Rank emblem + LP
- TFT: Shows off your TFT rank emblem + LP
- Arena: Shows off your Arena medallion + Your rating

If you want to hide your rank, then add the ``--no-rank`` argument, to **disable** this feature. As it's enabled by default.

![lobby-ranked](images/in_soloq_show_ranked_1.png) ![lobby-ranked-2](images/in_soloq_show_ranked_2.png)

### In Game
- Show your selected skin.
  - **Animated skins**: Ultimate skins will be animated on Discord.
  - **Skin Names**: The name of the skin will be shown when hovering the skin on Discord. This includes **Chromas** as well.
- **KDA**: Display your Kills, Deaths, Assists and Creep Score (cs)
  - Can be disabled with `--no-stats`
- **Rank**: Show what rank you have depending on the gamemode you play in (SoloQ, Flex, TFT, Arena, etc.)
  - Can be disabled with `--no-rank`
- **Game timer**: The ingame timer is accurately represented on Discord. Which is something even League's own Rich Presence don't do.


#### Skins
![Aphelios-skin](images/in_game_aphelios_skin_kda.png)


Example on Discord:

![Ezreal-Animated](images/animated_ezreal_showcase.gif) ![Lux-Animated](images/animated_lux_showcase.gif)

##### All Animated Skins

<div align="left">
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Ahri_86.gif" width="150" alt="Ahri"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Ezreal_5.gif" width="150" alt="Ezreal"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Jinx_60.gif" width="150" alt="Jinx"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Kaisa_71.gif" width="150" alt="Kaisa"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Lux_7.gif" width="150" alt="Lux"/>
</div>

<div align="left">
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/MissFortune_16.gif" width="150" alt="Miss Fortune"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Mordekaiser_54.gif" width="150" alt="Mordekaiser"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Morgana_80.gif" width="150" alt="Morgana"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Samira_30.gif" width="150" alt="Samira"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Seraphine_1.gif" width="150" alt="Seraphine"/>
</div>

<div align="left">
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Seraphine_2.gif" width="150" alt="Seraphine"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Seraphine_3.gif" width="150" alt="Seraphine"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Sett_66.gif" width="150" alt="Sett"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Sona_6.gif" width="150" alt="Sona"/>
  <img src="https://raw.githubusercontent.com/Its-Haze/league-assets/master/animated_skins/Udyr_3.gif" width="150" alt="Udyr"/>
</div>

### TFT (Teamfight tactics)
Showcase your favorite TFT Companion!

![tft-companion-1](images/tft_companion_showcase_1.png) ![tft-companion-2](images/tft_companion_showcase_2.png)

---

## Command Line Arguments

All arguments are optional - LeagueRPC works perfectly fine without any of them. Use these if you want to customize behavior.

**✨ = Enabled by default**

### `--launch-league <location>` ✨
LeagueRPC automatically finds and launches League for you. This is important because it takes priority over League's native Discord presence during startup.

Only specify a path if League is installed somewhere unusual:
```sh
leagueRPC.exe --launch-league "G:\Riot Games\Riot Client\RiotClientServices.exe"
```

### `--client-id <discord-app-id>` ✨
Want to show a different game name on Discord? Create an app at the [Discord Developer Portal](https://discord.com/developers/applications) and use its Application ID.

```sh
leagueRPC.exe --client-id 1230607224296968303
```
Fun options:
- **League of Kittens**: `1230607224296968303`
- **League of Linux**: `1185274747836174377`

### `--no-stats`
Hides your KDA and CS from Discord.
```sh
leagueRPC.exe --no-stats
```

### `--no-rank`
Hides your rank, LP, and emblem from Discord.
```sh
leagueRPC.exe --no-rank
```

### `--hide-emojis`
Removes the 🟢/🔴 emojis next to your Online/Away status.
```sh
leagueRPC.exe --hide-emojis
```

![Online](images/in_client_online_status.png) ![Away](images/in_client_away_status.png)

### `--hide-in-client`
Hides your Rich Presence when you're just sitting in the client. It'll show up again when you queue, enter champ select, or start a game.
```sh
leagueRPC.exe --hide-in-client
```

### `--add-process <process-name>`
Using a Discord alternative or modified client? Add its process name here. Find it in Task Manager.
```sh
leagueRPC.exe --add-process CustomDiscord AnotherProcess
```

### `--wait-for-league <seconds>` ✨
How long to wait for League to start before giving up. Default is `-1` (waits forever).
```sh
leagueRPC.exe --wait-for-league 30
```
*Mostly useful for legacy Linux setups with Lutris*

### `--wait-for-discord <seconds>` ✨
How long to wait for Discord to start. Default is `-1` (waits forever).
```sh
leagueRPC.exe --wait-for-discord 30
```

### Combining Arguments
Mix and match whatever you need:
```sh
leagueRPC.exe --client-id 1230607224296968303 --no-stats --hide-emojis
```

---

## 💡 Tips

### 🛡️ Windows Defender flagging it as a virus?
Just whitelist it. This happens because the exe isn't code-signed (not paying $100/year for that). The source code is public if you want to verify it's safe, or just build it yourself.

### 🖱️ Adding arguments without using a terminal
Don't like terminals? No problem:
1. Right-click `leagueRPC.exe` → Create shortcut
2. Right-click the shortcut → Properties
3. In the `Target` field, add your arguments after `leagueRPC.exe`
4. Double-click the shortcut to run

---

## ❓FAQ

### 🚫 Will this get my account banned?
Nope! It only uses Riot's local API (`127.0.0.1:2999`), which is completely safe. Vanguard won't care about it either since it doesn't modify any game files nor gives you an advantage in game.

### 🛡️ Is this a virus?
No. Some antivirus software might flag it because it's not code-signed (costs $100/year, not worth it for a free project). The entire source code is public on GitHub - feel free to review it or build it yourself. If you trust it, just whitelist it in Windows Defender.

### 🛠️ League's native RPC is still showing instead of LeagueRPC
Make sure LeagueRPC launches League for you. There's a tiny window during client startup where the native Discord presence can be disabled, and LeagueRPC needs to catch it.

If it's still not working:
1. Log out of League
2. Close League completely
3. Start LeagueRPC and let it launch League for you
4. Log back in

Still broken? Hit me up on [Discord](https://discord.haze.sh) or open a GitHub issue.

### ✔️ Does Riot approve this?
This is an independent open-source project, not affiliated with Riot Games.

### 🎮 Does it support TFT, Arena, ARAM, etc?
Yep! Works with all game modes including TFT, Arena, ARAM, Swarms, and whatever new modes Riot releases.

### 📉 Why doesn't my CS update live?
Blame Riot's API - it only updates every 10 minions killed instead of every single one. Nothing I can do about that unfortunately.

---

## 🏗️ Build from Source
For the cool kids who want to build it themselves:

```powershell
# Clone and navigate
git clone https://github.com/Its-Haze/league-rpc.git
cd league-rpc

# Set up virtual environment
python -m venv venv
.\venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt
pip install pyinstaller

# Build
pyinstaller --onefile --name leagueRPC.exe league_rpc/__main__.py --clean --distpath .

# Run
.\leagueRPC.exe
```

---

## 📞 Contact and Support
Got questions? Join the [Discord Server](https://discord.haze.sh) 
Feel free to open up Help tickets, or contact me directly on Discord (@haze.dev).

For issues related to the code, or project as a whole, please open an [issue on GitHub](https://github.com/Its-Haze/league-rpc/issues).

---

## 📜 History

This project was previously called `league-rpc-linux`, but since RIOT introduced Vanguard, and broke League on linux. I wanted to rename this project to `league-rpc`, and continue maintaining it for Windows users.

## Star History

<a href="https://star-history.com/#its-haze/league-rpc&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=its-haze/league-rpc&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=its-haze/league-rpc&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=its-haze/league-rpc&type=Date" />
 </picture>
</a>

