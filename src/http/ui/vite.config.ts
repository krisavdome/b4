import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const REMOTE_BACKEND =
  process.env.VITE_BACKEND_URL || "http://192.168.1.1:7000";

export default defineConfig({
  plugins: [react()],
  build: { outDir: "dist", emptyOutDir: true },
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: REMOTE_BACKEND,
        changeOrigin: true,
        ws: true,
        secure: false,
      },
    },
  },
});
