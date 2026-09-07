import { useReducer } from "react";
import type { Config } from "../../../bindings/github.com/its-haze/league-rpc/internal/config/models";
import { initialOnboarding, isFirstStep, isLastStep, onboardingReducer } from "../../lib/onboarding";
import { Button } from "../ui";
import { ClosingStep } from "./ClosingStep";
import { HowItWorksStep } from "./HowItWorksStep";
import { SettingsStep } from "./SettingsStep";
import { StartupStep } from "./StartupStep";
import { WelcomeStep } from "./WelcomeStep";

export interface OnboardingFlowProps {
  cfg: Config;
  applyPatch: (patch: Partial<Config>) => Promise<void>;
}

// Full-screen first-run flow, rendered instead of the app shell until Config.onboarding_complete is set.
export function OnboardingFlow({ cfg, applyPatch }: OnboardingFlowProps) {
  const [state, dispatch] = useReducer(onboardingReducer, initialOnboarding);

  function finish() {
    void applyPatch({ onboarding_complete: true });
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col items-center justify-center overflow-y-auto p-8">
      <div key={state.step} className="pane-enter flex w-full max-w-3xl flex-col gap-8">
        {state.step === "welcome" && <WelcomeStep />}
        {state.step === "settings" && <SettingsStep cfg={cfg} applyPatch={applyPatch} />}
        {state.step === "startup" && <StartupStep cfg={cfg} applyPatch={applyPatch} />}
        {state.step === "how-it-works" && <HowItWorksStep cfg={cfg} />}
        {state.step === "cta" && <ClosingStep />}

        <div className="flex items-center justify-between">
          <Button variant="ghost" onClick={() => dispatch({ type: "back" })} disabled={isFirstStep(state.step)}>
            Back
          </Button>
          {isLastStep(state.step) ? (
            <Button variant="primary" onClick={finish}>
              Let's go
            </Button>
          ) : (
            <Button variant="primary" onClick={() => dispatch({ type: "next" })}>
              Continue
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
