import { Radio } from "lucide-react";
import type { StatusSnapshot } from "../../../../bindings/github.com/its-haze/league-rpc/internal/app/models";
import { useDiscordAppName } from "../../../hooks/useDiscordAppName";
import { useSettings } from "../../../hooks/useSettings";
import { DiscordPresenceCard } from "../../DiscordPresenceCard";
import { SettingsCard } from "../../ui";

export interface PresencePreviewProps {
  status: StatusSnapshot | null;
}

// Mirrors what the Updater actually last sent to Discord, never a
// recomputation, so this can never disagree with what Discord shows.
export function PresencePreview({ status }: PresencePreviewProps) {
  const { cfg } = useSettings();
  const appName = useDiscordAppName(cfg?.discord_app_id ?? "");
  const presence = status?.presence;

  return (
    <SettingsCard
      icon={Radio}
      title="Discord presence"
      description="Exactly what League RPC last sent to Discord, not a re-guess of it."
    >

      {status?.presence_cleared || !presence ? (
        <p className="text-muted text-sm">Presence is currently cleared.</p>
      ) : (
        <DiscordPresenceCard
          details={presence.Details}
          state={presence.State}
          largeImage={presence.LargeImage}
          largeText={presence.LargeText}
          smallImage={presence.SmallImage}
          smallText={presence.SmallText}
          appName={appName}
          startUnixSeconds={presence.Start}
        />
      )}
    </SettingsCard>
  );
}
