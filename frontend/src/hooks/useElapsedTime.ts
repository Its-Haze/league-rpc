import { useEffect, useState } from "react";
import { formatElapsed } from "../lib/elapsed";

// Ticks every second and formats the time since startUnixSeconds, Discord-
// widget style ("2:05", "1:02:05"). Null while there's nothing to show.
export function useElapsedTime(startUnixSeconds: number): string | null {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  return formatElapsed(startUnixSeconds, now);
}
