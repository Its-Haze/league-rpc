import { useStatus } from "../../hooks/useStatus";
import { FeatureComparison } from "./home/FeatureComparison";
import { GithubCta } from "./home/GithubCta";
import { PresencePreview } from "./home/PresencePreview";
import { SupportCta } from "./home/SupportCta";

// The Home dashboard: the presence preview, what League RPC adds over native
// detection, and the two closing asks.
export function HomeScreen() {
  const status = useStatus();

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Home</h1>
      <PresencePreview status={status} />
      <FeatureComparison />
      <GithubCta />
      <SupportCta />
    </div>
  );
}
