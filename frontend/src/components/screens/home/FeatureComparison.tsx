import { Check, ListChecks, X } from "lucide-react";
import { SettingsCard } from "../../ui";

interface FeatureRow {
  label: string;
  native: boolean;
  leagueRpc: boolean;
}

// Native = Discord's own built-in "detected game" activity for League, with
const FEATURES: FeatureRow[] = [
  // Shared with native detection first, then league-rpc-only in priority order.
  { label: "Champion", native: true, leagueRpc: true },
  { label: "Skins, chromas & animated skins", native: false, leagueRpc: true },
  { label: "TFT companion", native: false, leagueRpc: true },
  { label: "KDA & CS", native: false, leagueRpc: true },
  { label: "Accurate in-game timer", native: false, leagueRpc: true },
  { label: "Ranked stats", native: false, leagueRpc: true },
  { label: "Custom presence text", native: false, leagueRpc: true },
  { label: "Summoner icons", native: false, leagueRpc: true },
  { label: "Spectating", native: false, leagueRpc: true },
];

function Mark({ on }: { on: boolean }) {
  return on ? (
    <Check className="text-ok size-4" aria-label="Yes" />
  ) : (
    <X className="text-muted size-4" aria-label="No" />
  );
}

export function FeatureComparison() {
  return (
    <SettingsCard
      icon={ListChecks}
      title="Why League RPC?"
      description="What Discord detects on its own, next to what League RPC adds."
    >

      <table className="w-full text-sm">
        <thead>
          <tr className="text-muted border-border border-b text-xs">
            <th className="py-1.5 text-left font-medium">Feature</th>
            <th className="w-20 py-1.5 text-center font-medium">Native</th>
            <th className="w-24 py-1.5 text-center font-medium">League RPC</th>
          </tr>
        </thead>
        <tbody>
          {FEATURES.map((row) => (
            <tr key={row.label} className="border-border-subtle border-b last:border-0">
              <td className="py-2 pr-4">{row.label}</td>
              <td className="py-2 text-center">
                <div className="flex justify-center">
                  <Mark on={row.native} />
                </div>
              </td>
              <td className="py-2 text-center">
                <div className="flex justify-center">
                  <Mark on={row.leagueRpc} />
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </SettingsCard>
  );
}
