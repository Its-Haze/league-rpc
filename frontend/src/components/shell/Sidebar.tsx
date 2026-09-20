import { Eye, Heart, House, Info, LifeBuoy, MessageCircleQuestion, SlidersHorizontal, Wrench, type LucideIcon } from "lucide-react";
import { Fragment, type ReactNode } from "react";
import { useStatus } from "../../hooks/useStatus";
import { useUpdateStatus } from "../../hooks/useUpdateStatus";
import { summarizeConnection, type ConnectionTone } from "../../lib/connectionStatus";
import { DiscordIcon, GitHubIcon } from "../icons";
import { DISCORD_COMMUNITY_URL, GITHUB_REPO_URL, SUPPORT_URL, openExternal } from "../../lib/links";
import { SECTIONS, type Section } from "../../lib/route";
import { type ThemeSetting } from "../../lib/theme";
import { ThemePicker } from "../ui";

const NAV: Record<Section, { label: string; icon: LucideIcon }> = {
  home: { label: "Home", icon: House },
  display: { label: "Display", icon: Eye },
  behavior: { label: "Behavior", icon: SlidersHorizontal },
  advanced: { label: "Advanced", icon: Wrench },
  faq: { label: "FAQ", icon: MessageCircleQuestion },
  help: { label: "Help", icon: LifeBuoy },
  about: { label: "About", icon: Info },
};

// Splits the nav into settings above and reference below. Named by the
// section it precedes, so reordering SECTIONS can't strand the rule.
const DIVIDER_BEFORE: Section = "faq";

const TONE_DOT: Record<ConnectionTone, string> = {
  ok: "bg-ok",
  warn: "bg-warn",
  idle: "bg-muted",
};

export interface SidebarProps {
  active: Section;
  onNavigate: (section: Section) => void;
  theme: string;
  onThemeChange: (theme: ThemeSetting) => void;
  /** True until the initial settings load resolves, so the picker can't fire on a config that isn't there yet. */
  themeDisabled?: boolean;
}

// Left nav across every section, plus a theme shortcut. Behavior has the
// labeled picker; this compact one is here for reaching it from anywhere.
export function Sidebar({ active, onNavigate, theme, onThemeChange, themeDisabled }: SidebarProps) {
  const connection = summarizeConnection(useStatus());
  const updateAvailable = useUpdateStatus()?.available ?? false;

  return (
    <nav className="border-border bg-surface flex w-44 shrink-0 flex-col border-r p-3">
      <div className="flex flex-1 flex-col gap-1">
        {SECTIONS.map((section) => {
          const isActive = section === active;
          const { label, icon: Icon } = NAV[section];
          const showUpdateDot = section === "about" && updateAvailable;
          return (
            <Fragment key={section}>
              {section === DIVIDER_BEFORE && <hr className="border-border my-2" />}
              <button
                onClick={() => onNavigate(section)}
                aria-current={isActive ? "page" : undefined}
                className={
                  "press flex items-center gap-2 rounded-sm px-2 py-1 text-left text-sm font-medium " +
                  (isActive
                    ? "bg-surface-raised text-text"
                    : "text-muted hover:bg-surface-raised hover:text-text")
                }
              >
                <span
                  className={
                    // Glyphs stay at full text weight on every row, so the icon
                    // column reads as a column; only the label dims when inactive.
                    "text-text relative grid size-8 shrink-0 place-items-center rounded-md transition-colors " +
                    (isActive ? "bg-accent-bg-strong" : "")
                  }
                >
                  <Icon className="size-5" />
                  {showUpdateDot && (
                    <span
                      key={section}
                      aria-hidden
                      className="bg-warn ring-surface update-dot-pop absolute top-0.5 right-0.5 size-2 rounded-full ring-2"
                    />
                  )}
                </span>
                {label}
                {showUpdateDot && <span className="sr-only">, update available</span>}
              </button>
            </Fragment>
          );
        })}
      </div>

      <button
        onClick={() => onNavigate("home")}
        title="Open Home for the full presence preview"
        className="text-muted hover:text-text press flex items-center gap-2 px-2 pb-2 text-xs"
      >
        {/* The ping only rides the healthy tone: it means presence is going out,
            and nothing is going out while paused or disconnected. */}
        <span
          className={
            "relative size-2 shrink-0 rounded-full " +
            TONE_DOT[connection.tone] +
            (connection.tone === "ok" ? " connection-ping" : "")
          }
        />
        {connection.label}
      </button>

      <div className="border-border flex flex-col gap-2 border-t pt-2">
        <SidebarLink href={DISCORD_COMMUNITY_URL} icon={<DiscordIcon className="size-5" />}>
          Discord
        </SidebarLink>
        <SidebarLink href={GITHUB_REPO_URL} icon={<GitHubIcon className="size-5" />}>
          GitHub
        </SidebarLink>
        <SidebarLink
          href={SUPPORT_URL}
          icon={
            <span className="heart-rise size-5">
              <Heart aria-hidden />
              <Heart className="rise-fill" fill="currentColor" aria-hidden />
            </span>
          }
        >
          Support
        </SidebarLink>
        <div className="mt-1">
          <ThemePicker value={theme} onChange={onThemeChange} disabled={themeDisabled} compact />
        </div>
      </div>
    </nav>
  );
}

function SidebarLink({ href, icon, children }: { href: string; icon: ReactNode; children: ReactNode }) {
  return (
    <a
      href={href}
      onClick={(e) => {
        e.preventDefault();
        openExternal(href);
      }}
      className="text-muted hover:bg-surface-raised hover:text-text press flex items-center gap-2 rounded-sm px-2 py-1 text-sm font-medium"
    >
      <span className="grid size-8 shrink-0 place-items-center">{icon}</span>
      {children}
    </a>
  );
}
