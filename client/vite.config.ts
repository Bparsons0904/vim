import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";
import path from "path";

export default defineConfig({
  envDir: "../",
  plugins: [solidPlugin()],
  server: {
    port: 3020,
    host: "0.0.0.0",
    hmr: {
      protocol: "ws",
      host: "localhost",
      port: 3020,
    },
    watch: {
      usePolling: false,
      ignored: ["**/.git/**", "**/node_modules/**", "**/dist/**"],
    },
  },
  test: {
    globals: false,
    environment: "jsdom",
    setupFiles: "./src/test/setup.ts",
    deps: {
      optimizer: {
        web: {
          include: ["solid-js", "@solidjs/testing-library"],
        },
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
      "@types": path.resolve(__dirname, "./src/types/"),
    },
    conditions: ["development", "browser"],
  },
  build: {
    target: "esnext",
  },
  css: {
    preprocessorOptions: {
      scss: {
        includePaths: [path.resolve(__dirname, "src/styles")],
        additionalData: `
          @use "@styles/variables" as *;
          @use "@styles/mixins" as *;
          @use "@styles/colors" as *;
        `,
      },
    },
  },
});
