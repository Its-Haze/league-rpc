---
name: ui-options
description: Propose two to four visual options for a UI element and let Haze pick from a throwaway side-by-side dark/light preview page in his browser, then implement the winner and delete the page. Use this whenever a UI decision has more than one reasonable answer - picking an icon, a status indicator, a badge, a card layout, table density, a label's wording or weight, spacing, an empty state, button hierarchy, or anything where "how should this look" is the real question. Reach for it proactively before building or restyling UI in this repo instead of silently picking one yourself, and whenever Haze says things like "show me some options", "what would look better", "I'm not sure how this should look", "make this look nicer", or asks you to redesign a component. Skip it for logic-only changes and for UI changes with one obvious answer.
---

# UI options preview

Picking how something looks is a judgement call, and describing two layouts in
prose almost never conveys the difference. This skill puts the real options on
screen, in both themes, rendered with the app's own Tailwind build and token
file, so the decision is made by looking rather than imagining.

## When it earns its keep

Use it when you catch yourself about to make an aesthetic choice on Haze's
behalf and more than one answer is defensible. A status row that could be a dot,
a pill, or a stacked block. An icon set where three lucide glyphs all fit. A
settings row that could be label-left or label-above.

Skip it when there's one obvious answer, when the change is behavioural, or when
Haze already told you exactly what he wants. Building a preview for a foregone
conclusion wastes his attention, which is the scarce resource here.

## The loop

1. Write `frontend/.preview/options.html` (structure below).
2. Run the preview:
   ```
   node .claude/skills/ui-options/scripts/preview.mjs --title "What's being decided"
   ```
   Run it in the background. It builds with the project's Vite and Tailwind,
   serves on `http://localhost:5199/`, and opens the browser. Add `--port N` if
   5199 is taken, `--build-only` to compile without serving.
3. Tell Haze it's open, name your recommendation and say in one sentence why.
   Then stop and wait. Do not start implementing on the assumption he'll agree.
4. He picks, one of two ways, both normal:
   - He clicks an option and hits "Send to Claude". That writes
     `frontend/.preview/choice.json` and the server exits, which ends the
     background task and wakes you. Read the file when it does:
     ```
     cat frontend/.preview/choice.json
     ```
     `picks` maps each decision id to an option letter; `notes` is free text and
     often matters more than the click ("B's shape but A's border weight").
     When the click and the notes disagree, the notes are the real instruction.
   - Or he just tells you in chat, which is the better channel for anything
     involved. Don't make him go back to the page to confirm a decision he has
     already stated.
5. Kill the server if it's still up, delete `frontend/.preview/`, implement the
   winner. The preview is scaffolding, and leaving it behind turns a clean repo
   into a confusing one.

## Writing options.html

The file is a fragment, not a whole page. The script wraps it in the shell,
generates the theme-scoped tokens from `frontend/src/tokens.css`, and clones
each `<template class="sample">` into a dark panel and a light panel, so you
write the markup once and it renders twice.

```html
<p class="preview-intro">One or two sentences on what's being decided.</p>

<section class="decision" data-decision="status-row">
  <h2>Connection status row</h2>
  <p class="ask">Which reads fastest at a glance in the tray-sized window?</p>
  <div class="options">
    <article class="option" data-option="A">
      <header>
        <span class="tag">A</span>
        <h3>Dot and label</h3>
        <span class="badge">Recommended</span>
      </header>
      <p class="rationale">Cheapest scan: colour carries the state, text confirms it.</p>
      <template class="sample">
        <!-- real component markup, real Tailwind classes -->
      </template>
    </article>
  </div>
</section>
```

`data-decision` is the key that shows up in `choice.json`, so make it a readable
slug. `data-option` is the letter. Repeat the `<section class="decision">` block
when one review covers several elements; each decision gets its own pick.

### Rules that keep the preview honest

**Use tokens, never hex.** Every colour goes through a Tailwind utility that
maps to `tokens.css`: `bg-surface`, `text-muted`, `border-accent-border`,
`bg-ok`, `text-danger`. A hardcoded colour looks fine in whichever theme you
designed it in and falls apart in the other, which defeats the whole point of
showing both. If an option genuinely needs a colour that isn't a token, say so
out loud - that's a signal the token set needs extending, and it's a separate
decision Haze should make deliberately.

**Show the element in context.** A button floating alone tells you nothing about
whether its weight is right. Wrap samples in enough surrounding structure - the
card it sits in, the row above it - that the comparison is about the real
situation. Keep the context identical across options so the only thing varying
is the thing being decided.

**Two to four options.** One isn't a choice. Five is a chore, and the weak ones
dilute the strong ones. If you have six ideas, pick the three you'd actually
ship and mention the others in chat.

**Recommend exactly one, with a reason.** Put `<span class="badge">Recommended</span>`
in its header and a one-line `rationale` on every option saying what it costs,
not just what it does. "More presence, but the tint competes with the accent
panels below it" is useful. "Clean and modern" is noise. Haze is choosing
between tradeoffs, so name them.

**Make the options actually different.** Three variations on the same idea with
2px of padding between them isn't a decision worth his time. Vary the approach:
shape, hierarchy, information density, whether a thing is text or colour.

## Cleanup

```
rm -rf frontend/.preview
```

Kill the background server first. `frontend/.preview/` is gitignored, but it
still clutters the working tree and confuses the next Vite build if it lingers.

## If the build fails

The script leans on the project's own `frontend/node_modules`, so a failure
usually means dependencies aren't installed - run `npm install` in `frontend/`.
If Tailwind classes come out unstyled, the class is probably one Tailwind can't
see: it only scans `frontend/.preview/index.html`, so classes assembled at
runtime in JavaScript won't be generated. Write them literally in the markup.
