import { describe, expect, it } from "vitest";
import { ONBOARDING_STEPS, initialOnboarding, isFirstStep, isLastStep, onboardingReducer } from "./onboarding";

describe("onboardingReducer", () => {
  it("starts on the first step", () => {
    expect(initialOnboarding.step).toBe(ONBOARDING_STEPS[0]);
  });

  it("advances through every step in order", () => {
    let state = initialOnboarding;
    for (let i = 1; i < ONBOARDING_STEPS.length; i++) {
      state = onboardingReducer(state, { type: "next" });
      expect(state.step).toBe(ONBOARDING_STEPS[i]);
    }
  });

  it("does not advance past the last step", () => {
    const last = { step: ONBOARDING_STEPS[ONBOARDING_STEPS.length - 1] };
    expect(onboardingReducer(last, { type: "next" })).toBe(last);
  });

  it("does not go back before the first step", () => {
    expect(onboardingReducer(initialOnboarding, { type: "back" })).toBe(initialOnboarding);
  });

  it("goes back a step", () => {
    const second = { step: ONBOARDING_STEPS[1] };
    expect(onboardingReducer(second, { type: "back" })).toEqual({ step: ONBOARDING_STEPS[0] });
  });

  it("jumps directly to a step", () => {
    expect(onboardingReducer(initialOnboarding, { type: "goto", step: "cta" })).toEqual({
      step: "cta",
    });
  });

  it("returns the same state object for a no-op goto", () => {
    expect(onboardingReducer(initialOnboarding, { type: "goto", step: initialOnboarding.step })).toBe(
      initialOnboarding,
    );
  });
});

describe("isFirstStep / isLastStep", () => {
  it("flags only the boundary steps", () => {
    expect(isFirstStep(ONBOARDING_STEPS[0])).toBe(true);
    expect(isLastStep(ONBOARDING_STEPS[0])).toBe(false);
    const last = ONBOARDING_STEPS[ONBOARDING_STEPS.length - 1];
    expect(isFirstStep(last)).toBe(false);
    expect(isLastStep(last)).toBe(true);
  });
});
