import { useEffect, useState } from "react";
import { GetDisplayPreview } from "../../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import type { Config } from "../../../bindings/github.com/its-haze/league-rpc/internal/config/models";
import { usePreviewAssets } from "../../hooks/usePreviewAssets";
import { withShowEmojis, withShowRank, withShowStats } from "../../lib/displayPatch";
import { DiscordPresenceCard } from "../DiscordPresenceCard";
import { Field, Toggle } from "../ui";

export interface SettingsStepProps {
  cfg: Config;
  applyPatch: (patch: Partial<Config>) => Promise<void>;
}

interface PreviewText {
  details: string;
  state: string;
}

const EMPTY_PREVIEW: PreviewText = { details: "", state: "" };

// Fixed elapsed baseline for the in-game demo card, so it reads as a match
// already in progress rather than one that just started.
const DEMO_ELAPSED_SECONDS = 14 * 60 + 32;

// Onboarding's settings screen. Two preview cards, not one: rank and stats
export function SettingsStep({ cfg, applyPatch }: SettingsStepProps) {
  const assets = usePreviewAssets();
  const showRank = cfg.display.default.show_rank;
  const showStats = cfg.display.default.show_stats;
  const showEmojis = cfg.presence.show_emojis;
  const inGameTemplate = cfg.presence.templates?.["in-game"] ?? { details: "", state: "" };
  const inClientTemplate = cfg.presence.templates?.["in-client"] ?? { details: "", state: "" };
  const [startUnixSeconds] = useState(() => Math.floor(Date.now() / 1000) - DEMO_ELAPSED_SECONDS);

  const [inGamePreview, setInGamePreview] = useState<PreviewText>(EMPTY_PREVIEW);
  const [inClientPreview, setInClientPreview] = useState<PreviewText>(EMPTY_PREVIEW);

  useEffect(() => {
    let cancelled = false;
    GetDisplayPreview("in-game", inGameTemplate, showStats, showEmojis)
      .then((p) => {
        if (!cancelled) setInGamePreview({ details: p.details, state: p.state });
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [inGameTemplate, showStats, showEmojis]);

  useEffect(() => {
    let cancelled = false;
    GetDisplayPreview("in-client", inClientTemplate, showStats, showEmojis)
      .then((p) => {
        if (!cancelled) setInClientPreview({ details: p.details, state: p.state });
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [inClientTemplate, showStats, showEmojis]);

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold">Pick what shows up</h1>
      <p className="text-muted text-base">
        These are the same toggles you'll find later under Display, where you can always change
        them again.
      </p>

      <section className="border-border bg-surface flex flex-col gap-1 rounded-lg border p-6">
        <Field id="onboarding-show-rank" label="Show rank" hint="Rank emblem and LP">
          <Toggle
            id="onboarding-show-rank"
            checked={showRank}
            onCheckedChange={(v) => void applyPatch(withShowRank(cfg, v))}
            label="Show rank"
          />
        </Field>
        <Field id="onboarding-show-stats" label="Show stats" hint="KDA and creep score">
          <Toggle
            id="onboarding-show-stats"
            checked={showStats}
            onCheckedChange={(v) => void applyPatch(withShowStats(cfg, v))}
            label="Show stats"
          />
        </Field>
        <Field id="onboarding-show-emojis" label="Show status emojis" hint="Online/away indicator">
          <Toggle
            id="onboarding-show-emojis"
            checked={showEmojis}
            onCheckedChange={(v) => void applyPatch(withShowEmojis(cfg, v))}
            label="Show status emojis"
          />
        </Field>
      </section>

      <div className="border-border flex flex-col gap-3 border-t pt-4">
        <div className="flex flex-col gap-2">
          <span className="text-muted text-xs font-semibold tracking-wide uppercase">In game</span>
          <DiscordPresenceCard
            details={inGamePreview.details}
            state={inGamePreview.state}
            largeImage={assets?.champion_skin_url}
            smallImage={showRank ? assets?.rank_emblem_url : assets?.league_logo_url}
            startUnixSeconds={startUnixSeconds}
          />
        </div>
        <div className="flex flex-col gap-2">
          <span className="text-muted text-xs font-semibold tracking-wide uppercase">In client</span>
          <DiscordPresenceCard
            details={inClientPreview.details}
            state={inClientPreview.state}
            largeImage={assets?.profile_icon_url}
            smallImage={assets?.league_logo_url}
          />
        </div>
      </div>
    </div>
  );
}
