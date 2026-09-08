// The presence contexts internal/presence/template knows about, display order.
export const PRESENCE_CONTEXTS = [
  "in-client",
  "lobby",
  "custom-lobby",
  "queue",
  "champ-select",
  "in-game",
  "tft-in-game",
  "spectating",
] as const;

export type PresenceContext = (typeof PRESENCE_CONTEXTS)[number];

export const PRESENCE_CONTEXT_LABELS: Record<PresenceContext, string> = {
  "in-client": "In client",
  lobby: "Lobby",
  "custom-lobby": "Custom game / practice tool",
  queue: "Queue",
  "champ-select": "Champ select",
  "in-game": "In game",
  "tft-in-game": "TFT in game",
  spectating: "Spectating",
};

export function isPresenceContext(value: string): value is PresenceContext {
  return (PRESENCE_CONTEXTS as readonly string[]).includes(value);
}
