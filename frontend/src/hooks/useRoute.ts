import { useCallback, useReducer } from "react";
import { initialRoute, routeReducer, type Section } from "../lib/route";

// Client-side routing between the sidebar sections. No history/URL sync:
// this is a single-window desktop app, so a full reload never happens anyway.
export function useRoute() {
  const [state, dispatch] = useReducer(routeReducer, initialRoute);
  const navigate = useCallback((section: Section) => dispatch({ type: "navigate", section }), []);
  return { section: state.section, navigate };
}
