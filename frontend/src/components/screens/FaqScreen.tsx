import { useState } from "react";
import { FAQ_GROUPS } from "../../lib/faq";
import { FaqItem } from "./faq/FaqItem";

// The FAQ section: questions grouped by topic, each one a collapsed row you
// open to read. Rows toggle independently, so comparing two answers works.
export function FaqScreen() {
  const [openIds, setOpenIds] = useState<ReadonlySet<string>>(new Set());

  function toggle(id: string) {
    setOpenIds((prev) => {
      const next = new Set(prev);
      if (!next.delete(id)) next.add(id);
      return next;
    });
  }

  return (
    <div className="flex flex-col gap-8">
      <h1 className="text-xl font-semibold">FAQ</h1>

      {FAQ_GROUPS.map((group) => (
        <section key={group.id} className="flex flex-col gap-3">
          <div className="flex items-center gap-2.5">
            <group.icon className="size-4 shrink-0" />
            <h2 className="text-muted text-xs font-semibold tracking-widest uppercase">{group.title}</h2>
          </div>
          <div className="flex flex-col gap-2">
            {group.entries.map((entry, i) => {
              // Positional, because a question doubles as an HTML id here and
              // the questions themselves carry spaces and punctuation.
              const id = `faq-${group.id}-${i}`;
              return (
                <FaqItem
                  key={entry.question}
                  id={id}
                  entry={entry}
                  open={openIds.has(id)}
                  onToggle={() => toggle(id)}
                />
              );
            })}
          </div>
        </section>
      ))}
    </div>
  );
}
