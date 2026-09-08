import { Plus } from "lucide-react";
import type { FaqEntry } from "../../../lib/faq";
import { openExternal } from "../../../lib/links";

export interface FaqItemProps {
  entry: FaqEntry;
  /** Unique across the page, so the button can point aria-controls at the answer. */
  id: string;
  open: boolean;
  onToggle: () => void;
}

// One question as a disclosure row: header always visible, answer sliding open
export function FaqItem({ entry, id, open, onToggle }: FaqItemProps) {
  return (
    <div
      className={
        "rounded-md border transition-colors " +
        (open ? "border-accent-border bg-surface-raised" : "border-border bg-surface hover:border-accent-border")
      }
    >
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={open}
        aria-controls={`${id}-answer`}
        className="flex w-full items-center gap-4 rounded-md px-4 py-3.5 text-left"
      >
        <span className="flex-1 text-sm font-medium">{entry.question}</span>
        <Plus
          aria-hidden
          className={
            "text-muted size-4 shrink-0 transition-transform duration-200 " + (open ? "rotate-45" : "")
          }
        />
      </button>

      {/* 0fr to 1fr is the one pure-CSS way to animate to a height nobody
          measured, so the answer stays mounted and both directions animate. */}
      <div
        id={`${id}-answer`}
        role="region"
        aria-hidden={!open}
        className={
          "grid transition-[grid-template-rows] duration-200 ease-out " +
          (open ? "grid-rows-[1fr]" : "grid-rows-[0fr]")
        }
      >
        <div className="overflow-hidden">
          <div className="flex max-w-[70ch] flex-col gap-2 px-4 pb-4">
            <p className="text-muted text-sm leading-relaxed">{entry.answer}</p>
            {entry.links && (
              <div className="flex flex-wrap gap-4">
                {entry.links.map((link) => (
                  <a
                    key={link.href}
                    href={link.href}
                    // Closed rows stay in the DOM for the animation, so their
                    // links have to be taken out of the tab order by hand.
                    tabIndex={open ? undefined : -1}
                    onClick={(e) => {
                      e.preventDefault();
                      openExternal(link.href);
                    }}
                    className="text-accent text-sm font-medium hover:underline"
                  >
                    {link.label}
                  </a>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
