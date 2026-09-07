import { AppWindow, EyeOff, Gamepad2, ShieldCheck, type LucideIcon } from "lucide-react";
import { DISCORD_COMMUNITY_URL, DISCORD_DEVELOPER_PORTAL_URL, GITHUB_REPO_URL } from "./links";

export interface FaqLink {
  label: string;
  href: string;
}

export interface FaqEntry {
  question: string;
  answer: string;
  links?: FaqLink[];
}

export interface FaqGroup {
  id: string;
  title: string;
  icon: LucideIcon;
  entries: FaqEntry[];
}

// Plain data, no JSX: the screen owns rendering, including routing links
// through openExternal so the webview doesn't navigate itself.
export const FAQ_GROUPS: FaqGroup[] = [
  {
    id: "safety",
    title: "Is this safe?",
    icon: ShieldCheck,
    entries: [
      {
        question: "Will this get my account banned?",
        answer:
          "No. It reads endpoints Riot already runs on your own machine, injects nothing and touches no game files. Vanguard has no reason to care.",
      },
      {
        question: "Why does Windows say it's dangerous?",
        answer:
          "It isn't code-signed, so SmartScreen gets twitchy about an .exe it hasn't seen before. That's the whole reason.",
        links: [{ label: "Read the source", href: GITHUB_REPO_URL }],
      },
      {
        question: "Is this made by Riot?",
        answer: "No. Independent open-source project, nothing to do with Riot.",
      },
    ],
  },
  {
    id: "not-showing",
    title: "Presence isn't showing",
    icon: EyeOff,
    entries: [
      {
        question: "League's own status shows instead",
        answer:
          "Give it a few seconds. Discord lets League's built-in integration overwrite ours, and League RPC keeps resending until it wins. Still wrong a minute later? That's a bug.",
        links: [{ label: "Ask on Discord", href: DISCORD_COMMUNITY_URL }],
      },
      {
        question: "Discord shows nothing at all",
        answer:
          "It only talks to Discord while League is running, so your status clears on purpose when you close the client. Check you're on the Discord desktop app too, the browser has no Rich Presence.",
      },
      {
        question: "Did I leave it paused?",
        answer:
          "Pause is on the Behavior page and it clears your status until you switch it back off. Worth a glance before assuming something's broken.",
      },
    ],
  },
  {
    id: "app",
    title: "Using the app",
    icon: AppWindow,
    entries: [
      {
        question: "I opened League RPC but League didn't start",
        answer:
          "It doesn't launch League, and that's deliberate: it's built to start with Windows, and nobody wants League opening the moment they boot. Start League however you normally do and League RPC picks it up, even mid-game.",
      },
      {
        question: "I closed the window and it's still running",
        answer:
          "On purpose. It hides to the tray so your presence keeps updating. Click the tray icon to bring the window back, right-click it for pause and quit, or make the X really mean quit under Behavior.",
      },
      {
        question: "Where are the logs?",
        answer:
          "Help page, Open logs folder. Copy diagnostics is faster and grabs everything I'd ask you for anyway.",
      },
      {
        question: "Can I change the app name Discord shows?",
        answer:
          "Advanced page. Presets switch between the Discord applications this ships with, Custom takes your own App ID. The bold name comes from the application, the lines under it from Display.",
        links: [{ label: "Discord developer portal", href: DISCORD_DEVELOPER_PORTAL_URL }],
      },
    ],
  },
  {
    id: "game-data",
    title: "Game data",
    icon: Gamepad2,
    entries: [
      {
        question: "Does it support TFT, Arena, ARAM?",
        answer:
          "All of them. Rift, ARAM, Arena, TFT with Double Up and Hyper Roll, plus whatever mode is rotating this month.",
      },
      {
        question: "Why does my CS jump instead of counting up?",
        answer: "Blame Riot's API. It only reports creep score in steps of ten, not per minion.",
      },
    ],
  },
];
