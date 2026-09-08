import type { Config } from "../../bindings/github.com/its-haze/league-rpc/internal/config/models";

// A representative Config tree for pure-logic tests; not the source of truth
// for defaults (internal/config.DefaultConfig() is), just a fixture shape.
export function DefaultConfig(): Config {
  return {
    schema_version: 1,
    discord_app_id: "1194034071588851783",
    theme: "system",
    onboarding_complete: false,
    display: {
      default: { show_rank: true, show_stats: true },
    },
    presence: {
      show_emojis: true,
      show_in_client: true,
      templates: {
        "in-client": { details: "{emoji}  {availability}", state: "In Client" },
        lobby: { details: "{queue}", state: "In Lobby ({players}/{max_players})" },
        "custom-lobby": { details: "{queue}", state: "In Lobby" },
        queue: { details: "{queue}", state: "In Queue" },
        "champ-select": { details: "{queue}", state: "In Champ Select" },
        "in-game": { details: "{queue}", state: "In Game · {stats}" },
        "tft-in-game": { details: "{queue}", state: "In Game · lvl: {level}" },
        spectating: { details: "{mode}", state: "Spectating" },
      },
    },
    behavior: {
      launch_at_startup: false,
      close_action: "ask",
      notify_updates: true,
    },
    advanced: {
      update_interval: 1500,
      stats_polling_interval: 3000,
      debug_mode: false,
    },
  };
}
