import { RotateCcw, type LucideIcon } from "lucide-react";
import type { ReactNode } from "react";

export interface SettingsCardProps {
  icon: LucideIcon;
  title: string;
  /** One line on what the card covers. Skip it when the title says it all. */
  description?: string;
  /** Control that belongs to the card itself, sitting on the header row: a
   * toggle for a single-setting card, or a button for an action card. */
  action?: ReactNode;
  /** Settings rows or content below the header. */
  children?: ReactNode;
  /** Draws the card in accent colours, for a setting worth recommending. */
  highlighted?: boolean;
  /** Short pill beside the title, e.g. "Recommended". */
  badge?: string;
  /** Reverts this card's setting to its built-in default. */
  onReset?: () => void;
  /** True once the value already matches the default, so reset can hide. */
  isDefault?: boolean;
}

// Titled card with an icon, the shape every settings section uses.
export function SettingsCard({ icon: Icon, title, description, action, children, highlighted, badge, onReset, isDefault }: SettingsCardProps) {
  return (
    <section
      className={
        "flex flex-col gap-3 rounded-lg border p-6 " +
        (highlighted ? "border-accent-border bg-accent-bg" : "border-border bg-surface")
      }
    >
      <div className="flex items-center justify-between gap-6">
        <div className="flex items-start gap-3">
          <Icon className="mt-0.5 size-5 shrink-0" />
          <div className="flex flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-base font-semibold">{title}</h2>
              {badge && (
                <span className="border-accent-border text-accent rounded-full border px-2 py-0.5 text-xs font-medium">
                  {badge}
                </span>
              )}
            </div>
            {description && <p className="text-muted text-sm">{description}</p>}
          </div>
        </div>
        {action && (
          // The reset slot is always reserved, so a card without one still
          // ends its control at the same x as every other card.
          <div className="flex shrink-0 items-center gap-2">
            <div className="flex items-center gap-3">{action}</div>
            <button
              type="button"
              onClick={onReset}
              title="Reset to default"
              aria-label={`Reset ${title} to default`}
              disabled={!onReset || isDefault}
              aria-hidden={!onReset}
              tabIndex={onReset ? undefined : -1}
              className="text-muted hover:text-text shrink-0 rounded-sm p-1 disabled:pointer-events-none disabled:opacity-0"
            >
              <RotateCcw className="size-3.5" />
            </button>
          </div>
        )}
      </div>
      {children && <div className="border-border flex flex-col gap-1 border-t pt-3">{children}</div>}
    </section>
  );
}
