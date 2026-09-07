// The sidebar sections. Order here is the order they render in the nav.
export const SECTIONS = ["home", "display", "behavior", "advanced", "faq", "help", "about"] as const;

export type Section = (typeof SECTIONS)[number];

export function isSection(value: string): value is Section {
  return (SECTIONS as readonly string[]).includes(value);
}

export interface RouteState {
  section: Section;
}

export type RouteAction = { type: "navigate"; section: Section };

export const initialRoute: RouteState = { section: "home" };

// Pure reducer so navigation is unit-testable without mounting React.
export function routeReducer(state: RouteState, action: RouteAction): RouteState {
  switch (action.type) {
    case "navigate":
      return state.section === action.section ? state : { section: action.section };
    default:
      return state;
  }
}
