import { Eye, MessageSquareText } from "lucide-react";
import type { TemplatePair } from "../../../bindings/github.com/its-haze/league-rpc/internal/config/models";
import { useDefaultConfig } from "../../hooks/useDefaultConfig";
import { useSettings } from "../../hooks/useSettings";
import { withShowEmojis, withShowInClient, withShowRank, withShowStats } from "../../lib/displayPatch";
import { PRESENCE_CONTEXT_LABELS, PRESENCE_CONTEXTS } from "../../lib/presenceContexts";
import { Field, SettingsCard, Tabs, Toggle } from "../ui";
import { TemplateEditor } from "./display/TemplateEditor";

// The Display section: global toggles, and one tab per presence context so
// the four text editors don't all show at once.
export function DisplayScreen() {
  const { cfg, error, applyPatch } = useSettings();
  const defaults = useDefaultConfig();

  if (!cfg) {
    return <p className="text-muted text-sm">Loading settings…</p>;
  }

  function setTemplate(ctx: string, next: TemplatePair) {
    void applyPatch({
      presence: { ...cfg!.presence, templates: { ...cfg!.presence.templates, [ctx]: next } },
    });
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Display</h1>
      {error && <p className="text-danger text-sm">{error}</p>}

      <SettingsCard
        icon={Eye}
        title="What your status shows"
        description="The extras League RPC adds on top of your champion and queue."
      >
        <Field
          id="show-rank"
          label="Show rank"
          hint="Rank emblem and LP"
          onReset={defaults ? () => void applyPatch(withShowRank(cfg, defaults.display.default.show_rank)) : undefined}
          isDefault={!defaults || cfg.display.default.show_rank === defaults.display.default.show_rank}
        >
          <Toggle
            id="show-rank"
            checked={cfg.display.default.show_rank}
            onCheckedChange={(v) => void applyPatch(withShowRank(cfg, v))}
            label="Show rank"
          />
        </Field>
        <Field
          id="show-stats"
          label="Show stats"
          hint="KDA and creep score"
          onReset={defaults ? () => void applyPatch(withShowStats(cfg, defaults.display.default.show_stats)) : undefined}
          isDefault={!defaults || cfg.display.default.show_stats === defaults.display.default.show_stats}
        >
          <Toggle
            id="show-stats"
            checked={cfg.display.default.show_stats}
            onCheckedChange={(v) => void applyPatch(withShowStats(cfg, v))}
            label="Show stats"
          />
        </Field>
        <Field
          id="show-emojis"
          label="Show status emojis"
          hint="Online/away indicator"
          onReset={defaults ? () => void applyPatch(withShowEmojis(cfg, defaults.presence.show_emojis)) : undefined}
          isDefault={!defaults || cfg.presence.show_emojis === defaults.presence.show_emojis}
        >
          <Toggle
            id="show-emojis"
            checked={cfg.presence.show_emojis}
            onCheckedChange={(v) => void applyPatch(withShowEmojis(cfg, v))}
            label="Show status emojis"
          />
        </Field>
        <Field
          id="show-in-client"
          label="Show presence while in client"
          hint="Keeps your status up between games, not only during one"
          onReset={defaults ? () => void applyPatch(withShowInClient(cfg, defaults.presence.show_in_client)) : undefined}
          isDefault={!defaults || cfg.presence.show_in_client === defaults.presence.show_in_client}
        >
          <Toggle
            id="show-in-client"
            checked={cfg.presence.show_in_client}
            onCheckedChange={(v) => void applyPatch(withShowInClient(cfg, v))}
            label="Show presence while in client"
          />
        </Field>
      </SettingsCard>

      <SettingsCard
        icon={MessageSquareText}
        title="Presence text"
        description="Write your own wording for each situation, or keep the defaults."
      >
        <Tabs
          defaultValue={PRESENCE_CONTEXTS[0]}
          items={PRESENCE_CONTEXTS.map((ctx) => {
            const pair = cfg.presence.templates?.[ctx] ?? { details: "", state: "" };
            return {
              value: ctx,
              label: PRESENCE_CONTEXT_LABELS[ctx],
              content: (
                <TemplateEditor
                  ctx={ctx}
                  value={pair}
                  onChange={(next) => setTemplate(ctx, next)}
                  showRank={cfg.display.default.show_rank}
                  showStats={cfg.display.default.show_stats}
                  showEmojis={cfg.presence.show_emojis}
                  defaultValue={defaults?.presence.templates?.[ctx]}
                />
              ),
            };
          })}
        />
      </SettingsCard>
    </div>
  );
}
