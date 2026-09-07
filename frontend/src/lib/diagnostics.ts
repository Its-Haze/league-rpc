// Wraps the backend's diagnostics text in a Markdown code fence so pasting
// straight into a GitHub issue renders as a preformatted block.
export function formatDiagnostics(backend: string): string {
  return "```\n" + backend.trim() + "\n```";
}
