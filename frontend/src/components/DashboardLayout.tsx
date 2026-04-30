import { useState, useEffect, useCallback } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import type { ComponentChildren } from "preact";
import { DashboardHeader } from "./DashboardHeader";
import { AuthPromptDialog } from "./AuthPromptDialog";
import { useAuth } from "../contexts/AuthContext";
import { Icon } from "./Icon";

const SIDEBAR_STORAGE_KEY = "sidebar-open";

function getStoredSidebarState(): boolean {
  try {
    const stored = localStorage.getItem(SIDEBAR_STORAGE_KEY);
    if (stored !== null) return stored === "true";
  } catch {}
  return true;
}

export function DashboardLayout({
  children,
}: {
  children: ComponentChildren;
}) {
  const { t } = useTranslation();
  const { url } = useLocation();
  const { isAuthenticated } = useAuth();
  const [showAuthPrompt, setShowAuthPrompt] = useState(false);

  const [sidebarOpen, setSidebarOpen] = useState(getStoredSidebarState);

  const toggleSidebar = useCallback(() => {
    setSidebarOpen((v) => {
      const next = !v;
      try { localStorage.setItem(SIDEBAR_STORAGE_KEY, String(next)); } catch {}
      return next;
    });
  }, []);

  // Close sidebar on mobile when navigating
  useEffect(() => {
    if (window.innerWidth < 1024) {
      setSidebarOpen(false);
    }
  }, [url]);

  return (
    <div class="relative flex lg:h-dvh w-full flex-col lg:overflow-hidden bg-background-light dark:bg-background-dark text-charcoal dark:text-white font-display antialiased">
      <DashboardHeader sidebarOpen={sidebarOpen} onToggleSidebar={toggleSidebar} />

      <div class="flex flex-1 w-full max-w-[1600px] mx-auto lg:overflow-hidden">
        {/* Mobile backdrop (desktop only) */}
        {sidebarOpen && (
          <div
            class="fixed inset-0 z-30 bg-black/40 hidden lg:hidden"
            onClick={toggleSidebar}
          />
        )}

        {/* Sidebar (desktop only) */}
        <aside
          class={`hidden lg:flex flex-col border-r border-border-light dark:border-border-dark shrink-0 transition-[width,opacity,transform] duration-200
            lg:relative lg:top-auto lg:bottom-auto lg:z-auto
            ${sidebarOpen
              ? "w-64 opacity-100 translate-x-0 overflow-y-auto overflow-x-hidden"
              : "w-0 opacity-0 lg:translate-x-0 overflow-hidden"
            }`}
          role="navigation"
          aria-label={t("dashboard.nav.sidebar")}
        >
          <div class="flex flex-col gap-6 px-6 pb-6 pt-4 min-w-[16rem]">
            <nav class="flex flex-col gap-1" aria-label={t("dashboard.nav.mainNav")}>
              <a
                class={`flex items-center gap-3 px-3 py-2 rounded-lg font-bold no-underline ${
                  url === "/"
                    ? "bg-primary/10 text-primary"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href="/"
                aria-current={url === "/" ? "page" : undefined}
              >
                <Icon name="home" />
                {t("dashboard.nav.mainFeed")}
              </a>
              <a
                class={`flex items-center gap-3 px-3 py-2 rounded-lg no-underline cursor-pointer ${
                  url === "/analytics"
                    ? "bg-primary/10 text-primary font-bold"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href={isAuthenticated ? "/analytics" : undefined}
                onClick={() => { if (!isAuthenticated) setShowAuthPrompt(true); }}
                aria-current={url === "/analytics" ? "page" : undefined}
              >
                <Icon name="analytics" />
                {t("dashboard.nav.analytics")}
              </a>
              <a
                class={`flex items-center gap-3 px-3 py-2 rounded-lg no-underline cursor-pointer ${
                  url === "/channels" || url === "/channels/explore"
                    ? "bg-primary/10 text-primary font-bold"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href={isAuthenticated ? "/channels" : undefined}
                onClick={() => { if (!isAuthenticated) setShowAuthPrompt(true); }}
                aria-current={url === "/channels" || url === "/channels/explore" ? "page" : undefined}
              >
                <Icon name="subscriptions" />
                {t("dashboard.nav.bottomChannels")}
              </a>
              <a
                class={`flex items-center gap-3 px-3 py-2 rounded-lg no-underline cursor-pointer ${
                  url === "/playlists" || url.startsWith("/playlists/")
                    ? "bg-primary/10 text-primary font-bold"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href={isAuthenticated ? "/playlists" : undefined}
                onClick={() => { if (!isAuthenticated) setShowAuthPrompt(true); }}
                aria-current={url === "/playlists" || url.startsWith("/playlists/") ? "page" : undefined}
              >
                <Icon name="playlist_play" />
                {t("dashboard.nav.playlists")}
              </a>

            </nav>
          </div>
        </aside>

        {/* Main Content */}
        <main class="flex-1 flex flex-col min-w-0 lg:overflow-y-auto overflow-x-hidden pb-14 lg:pb-0">
          {children}
        </main>
      </div>

      {/* Bottom Navigation (mobile only) */}
      <nav
        class="fixed bottom-0 left-0 right-0 z-50 flex lg:hidden items-center justify-around border-t border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark"
        aria-label={t("dashboard.nav.mainNav")}
      >
        <a
          href="/"
          onClick={(e) => {
            if (url === "/") {
              e.preventDefault();
              window.scrollTo({ top: 0, behavior: "smooth" });
            }
          }}
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors ${
            url === "/"
              ? "text-primary font-bold"
              : "text-text-muted-light dark:text-text-muted-dark"
          }`}
          aria-current={url === "/" ? "page" : undefined}
        >
          <Icon name="home" class="text-xl" />
          {t("dashboard.nav.bottomFeed")}
        </a>
        <a
          href={isAuthenticated ? "/channels" : undefined}
          onClick={() => { if (!isAuthenticated) setShowAuthPrompt(true); }}
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors cursor-pointer ${
            url === "/channels" || url === "/channels/explore"
              ? "text-primary font-bold"
              : "text-text-muted-light dark:text-text-muted-dark"
          }`}
          aria-current={url === "/channels" || url === "/channels/explore" ? "page" : undefined}
        >
          <Icon name="subscriptions" class="text-xl" />
          {t("dashboard.nav.bottomChannels")}
        </a>
        <a
          href={isAuthenticated ? "/playlists" : undefined}
          onClick={() => { if (!isAuthenticated) setShowAuthPrompt(true); }}
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors cursor-pointer ${
            url === "/playlists" || url.startsWith("/playlists/")
              ? "text-primary font-bold"
              : "text-text-muted-light dark:text-text-muted-dark"
          }`}
          aria-current={url === "/playlists" || url.startsWith("/playlists/") ? "page" : undefined}
        >
          <Icon name="playlist_play" class="text-xl" />
          {t("dashboard.nav.bottomPlaylists")}
        </a>
        <a
          href={isAuthenticated ? "/analytics" : undefined}
          onClick={() => { if (!isAuthenticated) setShowAuthPrompt(true); }}
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors cursor-pointer ${
            url === "/analytics"
              ? "text-primary font-bold"
              : "text-text-muted-light dark:text-text-muted-dark"
          }`}
          aria-current={url === "/analytics" ? "page" : undefined}
        >
          <Icon name="analytics" class="text-xl" />
          {t("dashboard.nav.bottomAnalytics")}
        </a>
        <a
          href={isAuthenticated ? "/profile" : undefined}
          onClick={() => {
            if (!isAuthenticated) { setShowAuthPrompt(true); return; }
            if (url === "/profile") {
              window.scrollTo({ top: 0, behavior: "smooth" });
            }
          }}
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors cursor-pointer ${
            url === "/profile"
              ? "text-primary font-bold"
              : "text-text-muted-light dark:text-text-muted-dark"
          }`}
          aria-current={url === "/profile" ? "page" : undefined}
        >
          <Icon name="person" class="text-xl" />
          {t("dashboard.nav.bottomProfile")}
        </a>
      </nav>

      <AuthPromptDialog
        open={showAuthPrompt}
        onClose={() => setShowAuthPrompt(false)}
      />
    </div>
  );
}
