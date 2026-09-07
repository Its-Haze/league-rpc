import { Star } from "lucide-react";
import { GitHubIcon } from "../../icons";
import { GITHUB_REPO_URL, openExternal } from "../../../lib/links";
import { Button } from "../../ui";

// Home's last card: the one ask we make of happy users, placed where it
// won't compete with anything functional above it.
export function GithubCta() {
  return (
    <section className="border-border bg-surface flex items-center justify-between gap-4 rounded-lg border p-6">
      <div className="flex items-center gap-4">
        <div className="bg-surface-raised border-border flex size-10 shrink-0 items-center justify-center rounded-full border">
          <GitHubIcon className="size-5" />
        </div>
        <div className="flex flex-col gap-0.5">
          <h2 className="text-sm font-semibold">Enjoying League RPC?</h2>
          <p className="text-muted text-xs">A star on GitHub helps other players find it.</p>
        </div>
      </div>
      <Button
        variant="secondary"
        asChild
        className="shrink-0"
      >
        <a
          href={GITHUB_REPO_URL}
          onClick={(e) => {
            e.preventDefault();
            openExternal(GITHUB_REPO_URL);
          }}
        >
          <span className="star-rise size-4">
            <Star aria-hidden />
            <Star className="star-solid" fill="currentColor" aria-hidden />
          </span>
          Star on GitHub
        </a>
      </Button>
    </section>
  );
}
