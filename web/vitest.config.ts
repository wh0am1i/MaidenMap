import { defineConfig, type UserConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "node:path";

// Cast required because @vitejs/plugin-react's vite type can diverge from
// vitest/config's nested vite version. Runtime behaviour is unaffected.
export default defineConfig({
  plugins: [react()] as UserConfig["plugins"],
  resolve: {
    alias: { "@": path.resolve(__dirname, "src") },
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./vitest.setup.ts"],
    css: true,
  },
});
