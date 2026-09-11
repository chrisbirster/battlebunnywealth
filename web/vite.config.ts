import stylex from "@stylexjs/unplugin";
import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

export default defineConfig({
  plugins: [
    stylex.vite({
      useCSSLayers: true,
      devMode: "full",
    }),
    solid(),
  ],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://127.0.0.1:8080",
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
});
