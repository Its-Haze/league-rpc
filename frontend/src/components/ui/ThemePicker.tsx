import { THEME_ICONS, THEME_LABELS, THEME_SETTINGS, type ThemeSetting } from "../../lib/theme";

export interface ThemePickerProps {
  value: string;
  onChange: (theme: ThemeSetting) => void;
  /** Icon-only equal-width segments, for the sidebar where the labels don't fit. */
  compact?: boolean;
  /** True until the initial settings load resolves, so a click can't fire on a config that isn't there yet. */
  disabled?: boolean;
}

// Segmented icon picker for the theme. Shared by onboarding, Behavior and the
// sidebar so the three never drift on what a theme is called or looks like.
export function ThemePicker({ value, onChange, compact, disabled }: ThemePickerProps) {
  return (
    <div
      role="radiogroup"
      aria-label="Theme"
      className={
        "border-border bg-surface-raised shrink-0 items-center gap-1 rounded-lg border p-1 " +
        (compact ? "flex w-full" : "inline-flex")
      }
    >
      {THEME_SETTINGS.map((setting) => {
        const Icon = THEME_ICONS[setting];
        const label = THEME_LABELS[setting];
        const active = value === setting;
        return (
          <button
            key={setting}
            type="button"
            role="radio"
            aria-checked={active}
            aria-label={compact ? label : undefined}
            title={compact ? label : undefined}
            disabled={disabled}
            onClick={() => onChange(setting)}
            className={
              "flex items-center justify-center rounded-md transition-colors " +
              (compact ? "h-7 flex-1 " : "gap-1.5 px-3 py-1.5 text-sm font-medium ") +
              (active ? "bg-surface " : "") +
              (disabled ? "text-disabled" : active ? "text-accent" : "text-muted hover:text-text")
            }
          >
            <Icon className="size-4 shrink-0" />
            {!compact && label}
          </button>
        );
      })}
    </div>
  );
}
