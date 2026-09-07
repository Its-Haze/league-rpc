import { Bell, Info } from "lucide-react";
import type { Config } from "../../../bindings/github.com/its-haze/league-rpc/internal/config/models";
import { howItWorksCopy } from "../../lib/howItWorks";

export interface HowItWorksStepProps {
  cfg: Config;
}

// Onboarding's fourth screen: the lifecycle, worded from the settings picked on
// the previous screen. Read-only on purpose, the toggles are one Back away.
export function HowItWorksStep({ cfg }: HowItWorksStepProps) {
  const { steps, updates } = howItWorksCopy({
    launchAtStartup: cfg.behavior.launch_at_startup,
    notifyUpdates: cfg.behavior.notify_updates,
  });

  return (
    <div className="flex flex-col gap-5 text-center">
      <h1 className="text-3xl font-semibold">How this works</h1>
      <p className="text-muted mx-auto max-w-2xl text-base">
        League RPC runs in the background and follows your client. Here's the whole loop.
      </p>

      <section className="border-border bg-surface flex flex-col gap-5 rounded-lg border p-6 text-left">
        {steps.map((step, i) => (
          <div key={step.title} className="flex gap-4">
            <div className="flex flex-col items-center">
              <span className="border-accent-border bg-accent-bg text-accent flex size-7 shrink-0 items-center justify-center rounded-full border text-sm font-semibold">
                {i + 1}
              </span>
              {i < steps.length - 1 && <span className="bg-border mt-1 w-px flex-1" />}
            </div>
            <div className="flex flex-col gap-1 pb-1">
              <h2 className="text-base font-semibold">{step.title}</h2>
              <p className="text-muted text-sm leading-relaxed">{step.body}</p>
            </div>
          </div>
        ))}
      </section>

      <section className="border-accent-border bg-accent-bg flex items-start gap-3 rounded-lg border p-5 text-left">
        <Info className="mt-0.5 size-5 shrink-0" />
        <div className="flex flex-col gap-1">
          <h2 className="text-base font-semibold">It doesn't launch League for you</h2>
          <p className="text-muted text-sm leading-relaxed">
            Opening League RPC won't open League. That's deliberate: it's built to start with
            Windows, and nobody wants League opening the moment they boot. Start League however you
            normally do, and League RPC will pick it up.
          </p>
        </div>
      </section>

      <div className="text-muted flex items-start gap-3 text-left text-sm">
        <Bell className="mt-0.5 size-4 shrink-0" />
        <p className="leading-relaxed">{updates}</p>
      </div>
    </div>
  );
}
