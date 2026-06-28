import { defineConfig, type Plugin } from "vite";
import preact from "@preact/preset-vite";
import tailwindcss from "@tailwindcss/vite";
import Icons from "unplugin-icons/vite";
import { VitePWA } from "vite-plugin-pwa";
import { ViteImageOptimizer } from "vite-plugin-image-optimizer";

/**
 * Emit a .webp sibling for every PNG/JPEG that ends up in dist/.
 * The original is kept (Open Graph / Twitter Card crawlers want PNG/JPEG),
 * but app code can reference the .webp via <img src="/foo.webp">.
 */
function generateWebP(): Plugin {
  return {
    name: "generate-webp",
    apply: "build",
    enforce: "post",
    async closeBundle() {
      const [{ default: sharp }, fs, path] = await Promise.all([
        import("sharp"),
        import("node:fs/promises"),
        import("node:path"),
      ]);
      const dir = "dist";
      const entries = await fs.readdir(dir, { recursive: true, withFileTypes: true }).catch(() => null);
      if (!entries) return;
      const targets = entries
        .filter((e) => e.isFile() && /\.(png|jpe?g)$/i.test(e.name))
        .map((e) => path.join(e.parentPath ?? dir, e.name));
      await Promise.all(
        targets.map(async (src) => {
          const out = src.replace(/\.(png|jpe?g)$/i, ".webp");
          try {
            await sharp(src).webp({ quality: 82 }).toFile(out);
          } catch {
            // skip — image-optimizer-modified files may already be encoded
          }
        }),
      );
    },
  };
}

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    port: 3000,
    strictPort: true,
    host: "0.0.0.0",
    allowedHosts: ["frontend"],
  },
  build: {
    target: "es2020",
    cssCodeSplit: true,
    reportCompressedSize: false,
    rollupOptions: {
      output: {
        manualChunks: {
          "vendor-preact": ["preact", "preact-iso", "preact/hooks", "preact/compat"],
          "vendor-swr": ["swr"],
          "vendor-i18n": ["i18next", "react-i18next"],
          "vendor-axios": ["axios"],
        },
      },
    },
  },
  plugins: [
    Icons({ compiler: "raw" }),
    tailwindcss(),
    ViteImageOptimizer(),
    generateWebP(),
    preact({
      prerender: {
        enabled: true,
        renderTarget: "#app",
        additionalPrerenderRoutes: ["/404", "/terms", "/privacy", "/about"],
        previewMiddlewareEnabled: false,
        previewMiddlewareFallback: "/404",
      },
    }),
    VitePWA({
      registerType: "autoUpdate",
      manifest: {
        name: "anti-yt",
        short_name: "anti-yt",
        description:
          "A free YouTube viewer that reclaims your time from the algorithm. Strict time controls, whitelist-based filtering, and distraction-free focus zones.",
        theme_color: "#1a1a1a",
        background_color: "#1a1a1a",
        display: "standalone",
        start_url: "/",
        icons: [
          {
            src: "/icon-192x192.svg",
            sizes: "192x192",
            type: "image/svg+xml",
            purpose: "any",
          },
          {
            src: "/icon-512x512.svg",
            sizes: "512x512",
            type: "image/svg+xml",
            purpose: "any",
          },
          {
            src: "/icon-512x512.svg",
            sizes: "512x512",
            type: "image/svg+xml",
            purpose: "maskable",
          },
        ],
      },
      workbox: {
        globPatterns: ["**/*.{js,css,html,svg,webp,woff2}"],
        globIgnores: ["**/about-preview.png", "**/explore-banner.png"],
        navigateFallbackDenylist: [/^\/api\//],
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/fonts\.googleapis\.com\/.*/i,
            handler: "CacheFirst",
            options: {
              cacheName: "google-fonts-cache",
              expiration: { maxEntries: 10, maxAgeSeconds: 60 * 60 * 24 * 365 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
          {
            urlPattern: /^https:\/\/fonts\.gstatic\.com\/.*/i,
            handler: "CacheFirst",
            options: {
              cacheName: "gstatic-fonts-cache",
              expiration: { maxEntries: 10, maxAgeSeconds: 60 * 60 * 24 * 365 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
          {
            urlPattern: /\/api\/.*/i,
            handler: "NetworkFirst",
            options: {
              cacheName: "api-cache",
              expiration: { maxEntries: 50, maxAgeSeconds: 60 * 5 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
        ],
      },
    }),
  ],
});
