import { defineConfig } from "vitest/config";

// Separate from vite.config.ts: the Wails plugin resolves the generated
// bindings directory, which the pure-logic tests here don't need.
export default defineConfig({
  test: {
    environment: "node",
    include: ["src/**/*.test.ts"],
  },
});
