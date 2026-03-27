import { defineConfig } from "vite";
import preact from "@preact/preset-vite";
import tailwindcss from "@tailwindcss/vite";
import Icons from "unplugin-icons/vite";

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    port: 3000,
    strictPort: true,
    host: "0.0.0.0",
    allowedHosts: ["frontend"],
  },
  plugins: [
    Icons({ compiler: "raw" }),
    tailwindcss(),
    preact({
      prerender: {
        enabled: true,
        renderTarget: "#app",
        additionalPrerenderRoutes: ["/404", "/terms", "/privacy"],
        previewMiddlewareEnabled: false,
        previewMiddlewareFallback: "/404",
      },
    }),
  ],
});
