// The Help screen's log viewer keeps at most this many lines in memory; older
// ones fall off the front, same as the backend ring buffer.
export const MAX_TAIL_LINES = 2000;

// Appends line to lines, dropping the oldest entries past MAX_TAIL_LINES.
export function appendLine(lines: string[], line: string): string[] {
  return appendLines(lines, [line]);
}

// Appends every line in newLines in one pass, so a burst of buffered lines
// costs one array copy instead of one per line, then drops past the cap.
export function appendLines(lines: string[], newLines: string[]): string[] {
  if (newLines.length === 0) return lines;
  const next = [...lines, ...newLines];
  return next.length > MAX_TAIL_LINES ? next.slice(next.length - MAX_TAIL_LINES) : next;
}

// Scroll-lock: true once the viewport is within tolerance px of the bottom.
// New lines only auto-scroll while this holds, else the user reading history gets yanked back down.
export function isScrolledToBottom(
  scrollTop: number,
  clientHeight: number,
  scrollHeight: number,
  tolerance = 24,
): boolean {
  return scrollHeight - scrollTop - clientHeight <= tolerance;
}
