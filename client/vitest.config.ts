import { defineConfig } from "vitest/config";
import solidPlugin from "vite-plugin-solid";
import path from "path";

export default defineConfig({
  plugins: [solidPlugin()],
  test: {
    environment: "jsdom",
    setupFiles: [path.resolve(__dirname, "./src/test/setup.ts")],
    deps: {
      optimizer: {
        web: {
          include: ["solid-js"],
        },
      },
    },
    globals: true,
    // Disable vite-plugin-solid auto jest-dom setup
    server: {
      deps: {
        external: [/solid-js/],
      },
    },
  },
  resolve: {
    alias: {
      "@styles": path.resolve(__dirname, "./src/styles"),
      "@components": path.resolve(__dirname, "./src/components"),
      "@layout": path.resolve(__dirname, "./src/components/layout/"),
      "@pages": path.resolve(__dirname, "./src/pages/"),
      "@hooks": path.resolve(__dirname, "./src/hooks/"),
      "@services": path.resolve(__dirname, "./src/services/"),
      "@context": path.resolve(__dirname, "./src/context/"),
    },
  },
});

