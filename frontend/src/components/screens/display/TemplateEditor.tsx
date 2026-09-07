import { useEffect, useRef, useState } from "react";
import {
  GetDisplayPreview,
  GetTemplateTokens,
} from "../../../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import type { PreviewAssets } from "../../../../bindings/github.com/its-haze/league-rpc/internal/app/models";
import type { TemplatePair } from "../../../../bindings/github.com/its-haze/league-rpc/internal/config/models";
import { useDebouncedValue } from "../../../hooks/useDebouncedValue";
import { usePreviewAssets } from "../../../hooks/usePreviewAssets";
import type { PresenceContext } from "../../../lib/presenceContexts";
import { DiscordPresenceCard } from "../../DiscordPresenceCard";
import { Field } from "../../ui";

// How long to wait after the last keystroke before persisting the template
// and re-running the preview, so typing doesn't save/round-trip every key.
const COMMIT_DELAY_MS = 400;

// Mirrors internal/discord/presence.go's image choices for each context, so
// the preview always shows the large/small pair a real send would use.
function previewImages(
  ctx: PresenceContext,
  showRank: boolean,
  assets: PreviewAssets | null,
): { largeImage?: string; smallImage?: string } {
  if (!assets) return {};
  switch (ctx) {
    case "in-client":
      // BuildInClientPresence never swaps in a rank emblem.
      return { largeImage: assets.profile_icon_url, smallImage: assets.league_logo_url };
    case "lobby":
      // BuildInLobbyPresence: map icon by default, rank emblem when shown.
      return {
        largeImage: assets.profile_icon_url,
        smallImage: showRank ? assets.rank_emblem_url : assets.map_icon_url,
      };
    case "custom-lobby":
      // BuildInCustomLobbyPresence: always the map icon, no rank swap.
      return { largeImage: assets.profile_icon_url, smallImage: assets.map_icon_url };
    case "queue":
      // BuildInQueuePresence: map icon by default, rank emblem when shown.
      return {
        largeImage: assets.profile_icon_url,
        smallImage: showRank ? assets.rank_emblem_url : assets.map_icon_url,
      };
    case "champ-select":
      // BuildInChampSelectPresence: map icon by default, rank emblem when shown.
      return {
        largeImage: assets.profile_icon_url,
        smallImage: showRank ? assets.rank_emblem_url : assets.map_icon_url,
      };
    case "in-game":
      // BuildInGamePresence: league logo by default, rank emblem when shown.
      return {
        largeImage: assets.champion_skin_url,
        smallImage: showRank ? assets.rank_emblem_url : assets.league_logo_url,
      };
    case "tft-in-game":
      // BuildTFTInGamePresence: TFT companion large image, league logo small
      return {
        largeImage: assets.tft_companion_url,
        smallImage: showRank ? assets.rank_emblem_url : assets.league_logo_url,
      };
    case "spectating":
      // BuildSpectatingPresence: always the league logo as the small image.
      return { largeImage: assets.champion_skin_url, smallImage: assets.league_logo_url };
  }
}

export interface TemplateEditorProps {
  ctx: PresenceContext;
  value: TemplatePair;
  onChange: (next: TemplatePair) => void;
  /** Current Display.Default/Presence toggles, so the preview honors them the
   * same way a real send would (stats line, emoji, rank emblem vs. the League logo). */
  showRank: boolean;
  showStats: boolean;
  showEmojis: boolean;
  /** The built-in details/state pair for ctx, for the per-line reset button.
   * Undefined while defaults haven't loaded yet, which just hides the button. */
  defaultValue?: TemplatePair;
}

// One presence context's editor: details/state text fields, a live preview
// rendered through the real template engine, and a token reference.
export function TemplateEditor({ ctx, value, onChange, showRank, showStats, showEmojis, defaultValue }: TemplateEditorProps) {
  const [tokens, setTokens] = useState<string[]>([]);
  const [preview, setPreview] = useState<{ details: string; state: string; warnings: string[] }>({
    details: "",
    state: "",
    warnings: [],
  });
  const assets = usePreviewAssets();

  // Local draft so typing feels instant; onChange (which persists to the
  // daemon) only fires once the draft settles, via the debounce below.
  const [draft, setDraft] = useState(value);
  // Tracks the last pair *we* committed, so the round-tripped echo of our
  // own write doesn't clobber whatever the user has typed since.
  const lastSent = useRef(value);
  useEffect(() => {
    if (value.details !== lastSent.current.details || value.state !== lastSent.current.state) {
      setDraft(value);
    }
  }, [value]);
  const debouncedDraft = useDebouncedValue(draft, COMMIT_DELAY_MS);
  const mounted = useRef(false);

  useEffect(() => {
    GetTemplateTokens(ctx)
      .then((t) => setTokens(t ?? []))
      .catch(() => {});
  }, [ctx]);

  useEffect(() => {
    if (!mounted.current) {
      mounted.current = true;
    } else {
      lastSent.current = debouncedDraft;
      onChange(debouncedDraft);
    }

    let cancelled = false;
    GetDisplayPreview(ctx, debouncedDraft, showStats, showEmojis)
      .then((p) => {
        if (!cancelled) setPreview({ details: p.details, state: p.state, warnings: p.warnings ?? [] });
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
    // onChange is expected to be stable enough per render; only ctx, the
  }, [ctx, debouncedDraft, showStats, showEmojis]);

  const { largeImage, smallImage } = previewImages(ctx, showRank, assets);

  return (
    <div className="flex flex-col gap-4">
      <div className="text-muted text-xs">
        Available tokens: {tokens.length > 0 ? tokens.map((t) => `{${t}}`).join(" ") : "none"}
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Field
          id={`${ctx}-details`}
          label="Details line"
          stacked
          onReset={defaultValue ? () => setDraft({ ...draft, details: defaultValue.details }) : undefined}
          isDefault={!defaultValue || draft.details === defaultValue.details}
        >
          <input
            id={`${ctx}-details`}
            value={draft.details}
            onChange={(e) => setDraft({ ...draft, details: e.target.value })}
            placeholder="blank uses the built-in default"
            className="border-border bg-surface-raised text-text w-full rounded-sm border px-3 py-1.5 text-sm"
          />
        </Field>
        <Field
          id={`${ctx}-state`}
          label="State line"
          stacked
          onReset={defaultValue ? () => setDraft({ ...draft, state: defaultValue.state }) : undefined}
          isDefault={!defaultValue || draft.state === defaultValue.state}
        >
          <input
            id={`${ctx}-state`}
            value={draft.state}
            onChange={(e) => setDraft({ ...draft, state: e.target.value })}
            placeholder="blank uses the built-in default"
            className="border-border bg-surface-raised text-text w-full rounded-sm border px-3 py-1.5 text-sm"
          />
        </Field>
      </div>

      <div className="border-border flex flex-col gap-2 border-t pt-4">
        <span className="text-muted text-xs font-semibold tracking-wide uppercase">Preview</span>
        <DiscordPresenceCard details={preview.details} state={preview.state} largeImage={largeImage} smallImage={smallImage} />
        {preview.warnings.length > 0 && (
          <ul className="text-danger text-xs">
            {preview.warnings.map((w) => (
              <li key={w}>{w}</li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
