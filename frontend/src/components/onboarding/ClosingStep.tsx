import { CheckCircle2 } from "lucide-react";

// Onboarding's last screen: a clean sign-off, not another ask. GitHub/Discord already had their moment on the welcome screen.
export function ClosingStep() {
  return (
    <div className="flex flex-col items-center gap-5 text-center">
      <CheckCircle2 className="size-14" />
      <h1 className="text-3xl font-semibold">You're all set</h1>
      <section className="border-border bg-surface flex w-full flex-col gap-2 rounded-lg border p-6">
        <p className="text-base">
          Set it up how you like. Close the window when you're done and it keeps running from the
          tray.
        </p>
        <p className="text-muted text-sm">Want this walkthrough again? Help &rarr; Replay walkthrough.</p>
      </section>
      <p className="text-muted text-sm">
        Built and maintained by haze. Thanks for giving it a try, and enjoy your games. 💜
      </p>
    </div>
  );
}
