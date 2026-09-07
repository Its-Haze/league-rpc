import { DownloadCloud, PanelBottom, Save, SlidersHorizontal } from "lucide-react";
import { DiscordIcon, GitHubIcon } from "../icons";
import { DISCORD_COMMUNITY_URL, GITHUB_REPO_URL, openExternal } from "../../lib/links";
import { Button } from "../ui";

// What the app does for you, rather than what it puts on your profile: the next
// screen already previews the presence itself.
const HIGHLIGHTS = [
  {
    icon: SlidersHorizontal,
    title: "A modern desktop interface",
    body: "Every option is a toggle or a field in here, next to a live preview of the status Discord will show.",
  },
  {
    icon: Save,
    title: "Settings that persist",
    body: "Your choices are written to disk the instant you make them, and they survive restarts and updates.",
  },
  {
    icon: DownloadCloud,
    title: "A built-in updater",
    body: "New versions are found and installed from inside the app. No trip to the releases page.",
  },
  {
    icon: PanelBottom,
    title: "An always-on tray app",
    body: "It sits in the tray and can start with Windows, so your status is live whether or not you remembered to open it.",
  },
];

export function WelcomeStep() {
  return (
    <div className="flex flex-col gap-5 text-center">
      <h1 className="text-3xl font-semibold">Welcome to League RPC</h1>
      <p className="text-muted mx-auto max-w-2xl text-base">
        Your League games, on your Discord profile. Set it up once here, and it takes care of
        itself from then on.
      </p>
      <div className="grid gap-3 text-left sm:grid-cols-2">
        {HIGHLIGHTS.map(({ icon: Icon, title, body }) => (
          <section key={title} className="border-border bg-surface flex flex-col gap-2 rounded-lg border p-5">
            <div className="flex items-center gap-2">
              <Icon className="size-5 shrink-0" />
              <h2 className="text-base font-semibold">{title}</h2>
            </div>
            <p className="text-muted text-sm">{body}</p>
          </section>
        ))}
      </div>
      <div className="flex items-center justify-center gap-3">
        <Button variant="secondary" asChild>
          <a
            href={GITHUB_REPO_URL}
            onClick={(e) => {
              e.preventDefault();
              openExternal(GITHUB_REPO_URL);
            }}
          >
            <GitHubIcon className="size-4" />
            Star on GitHub
          </a>
        </Button>
        <Button variant="secondary" asChild>
          <a
            href={DISCORD_COMMUNITY_URL}
            onClick={(e) => {
              e.preventDefault();
              openExternal(DISCORD_COMMUNITY_URL);
            }}
          >
            <DiscordIcon className="size-4" />
            Join the Discord
          </a>
        </Button>
      </div>
    </div>
  );
}
