// Formats an elapsed duration the way Discord's own presence widget does:
export function formatElapsed(startUnixSeconds: number, nowMs: number): string | null {
  if (!startUnixSeconds || startUnixSeconds <= 0) return null;

  const totalSeconds = Math.floor(nowMs / 1000) - startUnixSeconds;
  if (totalSeconds < 0) return null;

  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const pad = (n: number) => String(n).padStart(2, "0");

  return hours > 0 ? `${hours}:${pad(minutes)}:${pad(seconds)}` : `${minutes}:${pad(seconds)}`;
}
