import { AppWindow, Bug, Gauge, RotateCcw } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useConfigBounds } from "../../hooks/useConfigBounds";
import { useDefaultConfig } from "../../hooks/useDefaultConfig";
import { useDiscordAppName } from "../../hooks/useDiscordAppName";
import { usePresets } from "../../hooks/usePresets";
import { useSettings } from "../../hooks/useSettings";
import {
  CUSTOM_PRESET_VALUE,
  formatIntervalSeconds,
  isCustomAppId,
  resolveSelectValue,
  type Bounds,
} from "../../lib/advancedBounds";
import { DISCORD_DEVELOPER_PORTAL_URL, openExternal } from "../../lib/links";
import { DebouncedTextField, Field, Select, SettingsCard, Toggle } from "../ui";

// The Advanced section: Discord App ID (preset or custom), the two tuning
// intervals clamped to the config package's bounds, and debug logging.
export function AdvancedScreen() {
  const { cfg, error, applyPatch } = useSettings();
  const defaults = useDefaultConfig();
  const presets = usePresets();
  const bounds = useConfigBounds();
  // null means "not editing custom yet"; once the user types, even an empty
  // string sticks, so clearing the field to retype it doesn't snap back.
  const [customAppId, setCustomAppId] = useState<string | null>(null);
  // The user's own last custom id, independent of whatever preset is active
  const lastCustomAppIdRef = useRef<string>("");

  // Every hook must run before the "still loading" early return below, so
  // these fall back to blank/inert values until cfg resolves.
  const selectValue = cfg ? resolveSelectValue(cfg.discord_app_id, presets, customAppId !== null) : "";
  const customValue = customAppId ?? cfg?.discord_app_id ?? "";
  const resolvedName = useDiscordAppName(selectValue === CUSTOM_PRESET_VALUE ? customValue.trim() : "");

  useEffect(() => {
    if (cfg && isCustomAppId(cfg.discord_app_id, presets)) {
      lastCustomAppIdRef.current = cfg.discord_app_id;
    }
  }, [cfg?.discord_app_id, presets]);

  if (!cfg) {
    return <p className="text-muted text-sm">Loading settings…</p>;
  }

  const presetOptions = [
    ...Object.keys(presets).map((name) => ({ value: name, label: name })),
    { value: CUSTOM_PRESET_VALUE, label: "Custom" },
  ];
  const appIdInvalid = selectValue === CUSTOM_PRESET_VALUE && customValue.trim() === "";

  function handlePresetChange(value: string) {
    if (value === CUSTOM_PRESET_VALUE) {
      const restored = lastCustomAppIdRef.current || cfg!.discord_app_id;
      setCustomAppId(restored);
      // Actually switch back to it, not just pre-fill the text box: picking
      // Custom again should mean "use my custom id", live, right away.
      if (restored !== cfg!.discord_app_id) void applyPatch({ discord_app_id: restored });
      return;
    }
    setCustomAppId(null);
    const id = presets[value];
    if (id) void applyPatch({ discord_app_id: id });
  }

  function handleCustomCommit(value: string) {
    setCustomAppId(value);
    if (value.trim() !== "") void applyPatch({ discord_app_id: value.trim() });
  }

  function handleAppIdReset() {
    if (!defaults) return;
    setCustomAppId(null);
    if (defaults.discord_app_id !== cfg!.discord_app_id) void applyPatch({ discord_app_id: defaults.discord_app_id });
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Advanced</h1>
      {error && <p className="text-danger text-sm">{error}</p>}

      <SettingsCard
        icon={AppWindow}
        title="Discord application"
        description="Which app your presence appears under, including its name and icon on your profile."
      >
        <Field
          id="app-id-preset"
          label="Application"
          hint="Pick a preset, or point it at your own Discord app"
          onReset={defaults ? handleAppIdReset : undefined}
          isDefault={!defaults || cfg.discord_app_id === defaults.discord_app_id}
        >
          <Select
            value={selectValue}
            onValueChange={handlePresetChange}
            options={presetOptions}
            aria-label="Discord Application ID preset"
          />
        </Field>
        {selectValue === CUSTOM_PRESET_VALUE && (
          <>
            <Field id="app-id-custom" label="Application ID">
              <DebouncedTextField
                id="app-id-custom"
                value={customValue}
                onCommit={handleCustomCommit}
                className="border-border bg-surface-raised text-text rounded-sm border px-3 py-1.5 text-sm"
                aria-invalid={appIdInvalid}
              />
            </Field>
            {resolvedName && (
              <p className="text-muted text-right text-xs">
                Resolves to <span className="text-text font-medium">{resolvedName}</span>
              </p>
            )}
            <CustomAppIdTutorial />
          </>
        )}
        {appIdInvalid && <p className="text-danger text-xs">Discord Application ID must not be empty.</p>}
      </SettingsCard>

      <SettingsCard
        icon={Gauge}
        title="Update speed"
        description="How often League RPC refreshes what Discord shows. Faster feels more live, slower is lighter on your PC."
      >
        <div className="flex flex-col gap-4 pt-1">
          <IntervalSlider
            id="update-interval"
            label="Discord status"
            description="How quickly your status updates when your queue, rank, or game changes."
            bounds={bounds.updateInterval}
            value={cfg.advanced.update_interval}
            defaultValue={defaults?.advanced.update_interval}
            onCommit={(v) => void applyPatch({ advanced: { ...cfg.advanced, update_interval: v } })}
          />
          <IntervalSlider
            id="stats-interval"
            label="In-game stats"
            description="How often your KDA and creep score refresh while you're in a game."
            bounds={bounds.statsPollingInterval}
            value={cfg.advanced.stats_polling_interval}
            defaultValue={defaults?.advanced.stats_polling_interval}
            onCommit={(v) => void applyPatch({ advanced: { ...cfg.advanced, stats_polling_interval: v } })}
          />
        </div>
      </SettingsCard>

      <SettingsCard
        icon={Bug}
        title="Debug logging"
        description="Writes far more detail to the log file, for when you're chasing a problem. Takes effect immediately."
        onReset={
          defaults ? () => void applyPatch({ advanced: { ...cfg.advanced, debug_mode: defaults.advanced.debug_mode } }) : undefined
        }
        isDefault={!defaults || cfg.advanced.debug_mode === defaults.advanced.debug_mode}
        action={
          <Toggle
            id="debug-mode"
            checked={cfg.advanced.debug_mode}
            onCheckedChange={(v) => void applyPatch({ advanced: { ...cfg.advanced, debug_mode: v } })}
            label="Debug logging"
          />
        }
      />
    </div>
  );
}

// A collapsed-by-default walkthrough for getting a custom Discord
// Application ID, for users who don't already have one lying around.
function CustomAppIdTutorial() {
  return (
    <details className="text-muted text-xs">
      <summary className="text-muted hover:text-text select-none font-medium">
        How do I get a custom Application ID?
      </summary>
      <ol className="mt-2 list-decimal space-y-1 pl-4">
        <li>
          Open the{" "}
          <a
            href={DISCORD_DEVELOPER_PORTAL_URL}
            onClick={(e) => {
              e.preventDefault();
              openExternal(DISCORD_DEVELOPER_PORTAL_URL);
            }}
            className="text-accent underline"
          >
            Discord Developer Portal
          </a>{" "}
          and sign in with your Discord account.
        </li>
        <li>Click "New Application", give it a name, and create it.</li>
        <li>On the app's "General Information" page, copy the "Application ID".</li>
        <li>Paste that ID into the field above.</li>
      </ol>
    </details>
  );
}

interface IntervalSliderProps {
  id: string;
  label: string;
  description: string;
  bounds: Bounds;
  value: number;
  /** The built-in default in ms, for the reset button. Undefined while
   * defaults haven't loaded yet, which just hides the button. */
  defaultValue?: number;
  onCommit: (ms: number) => void;
}

// A plain-language slider for a millisecond interval: shows seconds instead
// of raw milliseconds, and only persists once the drag or keypress ends.
function IntervalSlider({ id, label, description, bounds, value, defaultValue, onCommit }: IntervalSliderProps) {
  const [draft, setDraft] = useState(value);

  useEffect(() => setDraft(value), [value]);

  function commit() {
    if (draft !== value) onCommit(draft);
  }

  function reset() {
    if (defaultValue === undefined) return;
    setDraft(defaultValue);
    if (defaultValue !== value) onCommit(defaultValue);
  }

  const isDefault = defaultValue === undefined || draft === defaultValue;

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between gap-4">
        <label htmlFor={id} className="text-sm">
          {label}
        </label>
        <div className="flex items-center gap-2">
          <span className="text-accent text-sm font-medium">{formatIntervalSeconds(draft)}</span>
          <button
            type="button"
            onClick={reset}
            title="Reset to default"
            aria-label={`Reset ${label} to default`}
            disabled={isDefault}
            className="text-muted hover:text-text shrink-0 rounded-sm p-1 disabled:pointer-events-none disabled:opacity-0"
          >
            <RotateCcw className="size-3.5" />
          </button>
        </div>
      </div>
      <p className="text-muted text-xs">{description}</p>
      <input
        id={id}
        type="range"
        min={bounds.min}
        max={bounds.max}
        step={100}
        value={draft}
        onChange={(e) => setDraft(Number(e.target.value))}
        onPointerUp={commit}
        onKeyUp={commit}
        className="accent-accent w-full"
      />
      <div className="text-muted flex justify-between text-[11px]">
        <span>Faster</span>
        <span>Slower</span>
      </div>
    </div>
  );
}
