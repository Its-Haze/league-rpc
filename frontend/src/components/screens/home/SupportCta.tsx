import { Coffee, Heart } from "lucide-react";
import { SUPPORT_URL, openExternal } from "../../../lib/links";
import { Button } from "../../ui";

// The paid ask, kept separate from the star so the free one never reads as
// the lesser half of a single pitch.
export function SupportCta() {
  return (
    <section className="border-border bg-surface flex items-center justify-between gap-4 rounded-lg border p-6">
      <div className="flex items-center gap-4">
        <div className="bg-surface-raised border-border flex size-10 shrink-0 items-center justify-center rounded-full border">
          <Coffee className="size-5" />
        </div>
        <div className="flex flex-col gap-0.5">
          <h2 className="text-sm font-semibold">Support the project</h2>
          <p className="text-muted text-xs">
            No ads, no paywall, ever. Tips keep it maintained, and supporters get a role in the
            Discord.
          </p>
        </div>
      </div>
      <Button variant="secondary" asChild className="shrink-0">
        <a
          href={SUPPORT_URL}
          onClick={(e) => {
            e.preventDefault();
            openExternal(SUPPORT_URL);
          }}
        >
          <span className="heart-rise size-4">
            <Heart aria-hidden />
            <Heart className="rise-fill" fill="currentColor" aria-hidden />
          </span>
          Support
        </a>
      </Button>
    </section>
  );
}
