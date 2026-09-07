import { Bell, Palette, Power } from "lucide-react";
import * as LabelPrimitive from "@radix-ui/react-label";
import type { Config } from "../../../bindings/github.com/its-haze/league-rpc/internal/config/models";
import { withLaunchAtStartup, withNotifyUpdates } from "../../lib/behaviorPatch";
import { Toggle, ThemePicker } from "../ui";

export interface StartupStepProps {
  cfg: Config;
  applyPatch: (patch: Partial<Config>) => Promise<void>;
}

// Onboarding's third screen: settings about the app itself, not the presence it sends.
export function StartupStep({ cfg, applyPatch }: StartupStepProps) {
  return (
    <div className="flex flex-col gap-5 text-center">
      <h1 className="text-3xl font-semibold">Make it yours</h1>
      <p className="text-muted text-base">
        Three settings for the app itself, not the status it sends.
      </p>

      <section className="border-border bg-surface flex items-center justify-between gap-6 rounded-lg border p-5 text-left">
        <div className="flex items-start gap-3">
          <Palette className="mt-0.5 size-5 shrink-0" />
          <div className="flex flex-col gap-1">
            <h2 className="text-base font-semibold">Appearance</h2>
            <p className="text-muted text-sm">System follows whatever Windows is set to.</p>
          </div>
        </div>
        <ThemePicker value={cfg.theme} onChange={(t) => void applyPatch({ theme: t })} />
      </section>

      <section className="border-accent-border bg-accent-bg flex items-center justify-between gap-6 rounded-lg border p-5 text-left">
        <div className="flex items-start gap-3">
          <Power className="mt-0.5 size-5 shrink-0" />
          <div className="flex flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              <LabelPrimitive.Root htmlFor="onboarding-launch-at-startup" className="text-base font-semibold">
                Start with Windows
              </LabelPrimitive.Root>
              <span className="border-accent-border text-accent rounded-full border px-2 py-0.5 text-xs font-medium">
                Recommended
              </span>
            </div>
            <p className="text-muted text-sm">
              Launches minimized to the tray, so your status is live before your first game.
            </p>
          </div>
        </div>
        <div className="shrink-0">
          <Toggle
            id="onboarding-launch-at-startup"
            checked={cfg.behavior.launch_at_startup}
            onCheckedChange={(v) => void applyPatch(withLaunchAtStartup(cfg, v))}
            label="Start with Windows"
          />
        </div>
      </section>

      <section className="border-accent-border bg-accent-bg flex items-center justify-between gap-6 rounded-lg border p-5 text-left">
        <div className="flex items-start gap-3">
          <Bell className="mt-0.5 size-5 shrink-0" />
          <div className="flex flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              <LabelPrimitive.Root htmlFor="onboarding-notify-updates" className="text-base font-semibold">
                Update notifications
              </LabelPrimitive.Root>
              <span className="border-accent-border text-accent rounded-full border px-2 py-0.5 text-xs font-medium">
                Recommended
              </span>
            </div>
            <p className="text-muted text-sm">
              League changes constantly, and that's usually what breaks a companion app like this
              one. Get notified the moment a fix is ready, so you're never stuck running a version
              League has already outgrown. Not your thing? New releases are also announced in the
              Discord server.
            </p>
          </div>
        </div>
        <div className="shrink-0">
          <Toggle
            id="onboarding-notify-updates"
            checked={cfg.behavior.notify_updates}
            onCheckedChange={(v) => void applyPatch(withNotifyUpdates(cfg, v))}
            label="Update notifications"
          />
        </div>
      </section>
    </div>
  );
}
