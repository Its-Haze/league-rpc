import { useStatus } from "../../hooks/useStatus";
import { FeatureComparison } from "./home/FeatureComparison";
import { GithubCta } from "./home/GithubCta";
import { PresencePreview } from "./home/PresencePreview";

// The Home dashboard: the last-sent presence preview, a rundown of what
// League RPC adds over native detection, and a closing GitHub star ask.
export function HomeScreen() {
  const status = useStatus();

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Home</h1>
      <PresencePreview status={status} />
      <FeatureComparison />
      <GithubCta />
    </div>
  );
}
