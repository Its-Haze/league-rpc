// Fallback bounds, used only until GetConfigBounds() resolves (useConfigBounds).
export const UPDATE_INTERVAL_BOUNDS = { min: 500, max: 10000 };
export const STATS_POLLING_INTERVAL_BOUNDS = { min: 1000, max: 30000 };

export interface Bounds {
  min: number;
  max: number;
}

// Clamps value into bounds. A non-numeric input (NaN, from an empty or
// partially-typed field) falls back to the nearer bound's min.
export function clampToBounds(value: number, bounds: Bounds): number {
  if (Number.isNaN(value)) return bounds.min;
  return Math.min(bounds.max, Math.max(bounds.min, value));
}

// Formats a millisecond duration as a plain seconds label for non-technical
// users, e.g. 1500 -> "1.5s", 10000 -> "10s". At most one decimal place.
export function formatIntervalSeconds(ms: number): string {
  const seconds = Math.round(ms / 100) / 10;
  return `${seconds}s`;
}

// True when appId matches one of the named presets, so the picker can show
// the preset's name instead of falling into "custom".
export function presetNameFor(
  appId: string,
  presets: Record<string, string | undefined>,
): string | null {
  for (const [name, id] of Object.entries(presets)) {
    if (id === appId) return name;
  }
  return null;
}

export const CUSTOM_PRESET_VALUE = "__custom__";

// The value the preset <Select> should show: the matching preset's name, or
// the sentinel that switches the UI into custom-entry mode.
export function presetSelectValue(
  appId: string,
  presets: Record<string, string | undefined>,
): string {
  return presetNameFor(appId, presets) ?? CUSTOM_PRESET_VALUE;
}

// What the preset <Select> should actually show. editingCustom wins outright,
// even if appId still happens to equal a preset's id.
export function resolveSelectValue(
  appId: string,
  presets: Record<string, string | undefined>,
  editingCustom: boolean,
): string {
  return editingCustom ? CUSTOM_PRESET_VALUE : presetSelectValue(appId, presets);
}

// True when appId does not match any named preset, i.e. it's a custom id.
export function isCustomAppId(appId: string, presets: Record<string, string | undefined>): boolean {
  return presetNameFor(appId, presets) === null;
}
