import { useElapsedTime } from "../hooks/useElapsedTime";
import { GameControllerIcon } from "./icons";

export interface DiscordPresenceCardProps {
  details: string;
  state: string;
  largeImage?: string;
  largeText?: string;
  smallImage?: string;
  smallText?: string;
  /** The Discord Application's own name, shown as the card's bold title
   * exactly like Discord itself renders it. Omitted if not yet resolved. */
  appName?: string | null;
  /** Unix seconds. Omitted (or 0) hides the elapsed-time row, for a static
   * preview that isn't a real running presence. */
  startUnixSeconds?: number;
}

// Mirrors the actual Discord Rich Presence widget's layout: a small activity
export function DiscordPresenceCard({
  details,
  state,
  largeImage,
  largeText,
  smallImage,
  smallText,
  appName,
  startUnixSeconds,
}: DiscordPresenceCardProps) {
  const elapsed = useElapsedTime(startUnixSeconds ?? 0);

  return (
    <div className="bg-surface-raised flex gap-3 rounded-md p-3">
      {(largeImage || smallImage) && (
        <div className="relative size-16 shrink-0">
          {largeImage && (
            <img src={largeImage} alt="" title={largeText || undefined} className="size-16 rounded-md object-cover" />
          )}
          {smallImage && (
            <img
              src={smallImage}
              alt=""
              title={smallText || undefined}
              className="bg-surface-raised absolute -right-1.5 -bottom-1.5 size-6 rounded-full object-cover ring-4 ring-[var(--color-surface-raised)]"
            />
          )}
        </div>
      )}
      <div className="flex min-w-0 flex-col justify-center gap-0.5">
        <div className="text-muted text-xs">Playing</div>
        {appName && <div className="truncate text-sm font-semibold">{appName}</div>}
        {details && <div className="truncate text-sm">{details}</div>}
        {state && <div className="text-muted truncate text-sm">{state}</div>}
        {elapsed && (
          <div className="text-muted mt-1 flex items-center gap-1 text-xs">
            <GameControllerIcon className="size-3" />
            <span>{elapsed} elapsed</span>
          </div>
        )}
      </div>
    </div>
  );
}
