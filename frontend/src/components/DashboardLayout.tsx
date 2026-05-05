import { useState, useCallback } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import type { ComponentChildren } from "preact";
import { DashboardHeader } from "./DashboardHeader";
import { AuthPromptDialog } from "./AuthPromptDialog";
import { useAuth } from "../contexts/AuthContext";
import { Icon } from "./Icon";
import { REPORT_FORM_URL } from "../constants";

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

  const navItemClass = (active: boolean) =>
    `flex items-center gap-3 px-3 py-2 rounded-lg no-underline cursor-pointer transition-colors ${
      active
        ? "bg-primary/10 text-primary font-bold"
        : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium"
    }`;

  return (
    <div class="relative flex min-h-dvh tablet:h-dvh w-full flex-col tablet:overflow-hidden bg-background-light dark:bg-background-dark text-charcoal dark:text-white font-display antialiased">
      <DashboardHeader sidebarOpen={sidebarOpen} onToggleSidebar={toggleSidebar} />

      <div class="flex flex-1 w-full tablet:overflow-hidden">
        {/* Sidebar (tablet and up): full when open, fully hidden when closed */}
        <aside
          class={`hidden tablet:flex flex-col border-r border-border-light dark:border-border-dark shrink-0 transition-[width,opacity] duration-200 ${
            sidebarOpen
              ? "w-64 opacity-100 overflow-hidden"
              : "w-0 opacity-0 overflow-hidden"
          }`}
          role="navigation"
          aria-label={t("dashboard.nav.sidebar")}
          aria-hidden={!sidebarOpen}
        >
          <div class="flex flex-col overflow-y-auto flex-1 min-w-64">
            <nav class="flex flex-col gap-1 p-2 pt-4" aria-label={t("dashboard.nav.mainNav")}>
              <a
                class={navItemClass(url === "/")}
                href="/"
                aria-current={url === "/" ? "page" : undefined}
              >
                <Icon name="home" class="shrink-0 text-xl" />
                <span class="truncate text-sm">{t("dashboard.nav.mainFeed")}</span>
              </a>
              <a
                class={navItemClass(url === "/analytics")}
                href={isAuthenticated ? "/analytics" : undefined}
                onClick={() => { if (!isAuthenticated) setShowAuthPrompt(true); }}
                aria-current={url === "/analytics" ? "page" : undefined}
              >
                <Icon name="analytics" class="shrink-0 text-xl" />
                <span class="truncate text-sm">{t("dashboard.nav.analytics")}</span>
              </a>
              <a
                class={navItemClass(url === "/channels" || url === "/channels/explore")}
                href={isAuthenticated ? "/channels" : undefined}
                onClick={() => { if (!isAuthenticated) setShowAuthPrompt(true); }}
                aria-current={url === "/channels" || url === "/channels/explore" ? "page" : undefined}
              >
                <Icon name="subscriptions" class="shrink-0 text-xl" />
                <span class="truncate text-sm">{t("dashboard.nav.bottomChannels")}</span>
              </a>
              <a
                class={navItemClass(url === "/playlists" || url.startsWith("/playlists/"))}
                href={isAuthenticated ? "/playlists" : undefined}
                onClick={() => { if (!isAuthenticated) setShowAuthPrompt(true); }}
                aria-current={url === "/playlists" || url.startsWith("/playlists/") ? "page" : undefined}
              >
                <Icon name="playlist_play" class="shrink-0 text-xl" />
                <span class="truncate text-sm">{t("dashboard.nav.playlists")}</span>
              </a>
            </nav>
          </div>

          <div class="shrink-0 border-t border-border-light dark:border-border-dark px-2 min-w-64">
            <div class="pt-2 pb-1">
              <a
                href={REPORT_FORM_URL}
                target="_blank"
                rel="noopener noreferrer"
                class={navItemClass(false)}
              >
                <Icon name="flag" class="shrink-0 text-xl" />
                <span class="truncate text-sm">{t("profile.nav.reportProblem")}</span>
              </a>
            </div>
            <div class="border-t border-border-light dark:border-border-dark mx-1 my-1" />
            <div class="px-3 pt-2 pb-1">
              <a href="/about" class="text-xs text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white no-underline transition-colors">
                {t("dashboard.nav.aboutAntiYt")}
              </a>
            </div>
            <div class="px-3 pb-2 flex flex-wrap gap-x-3 gap-y-2">
              <a href="/terms" class="text-xs text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white no-underline transition-colors">
                {t("common.termsLink")}
              </a>
              <a href="/privacy" class="text-xs text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white no-underline transition-colors">
                {t("common.privacyPolicyLink")}
              </a>
            </div>
            <p class="px-3 pb-4 text-xs text-text-muted-light dark:text-text-muted-dark">
              {t("dashboard.sidebarFooter.copyright")}
            </p>
          </div>
        </aside>

        {/* Main Content */}
        <main class="flex-1 flex flex-col min-w-0 tablet:overflow-y-auto overflow-x-hidden pb-14 tablet:pb-0">
          {children}
        </main>
      </div>

      {/* Bottom Navigation (below 657px only) */}
      <nav
        class="fixed bottom-0 left-0 right-0 z-50 flex tablet:hidden items-center justify-around border-t border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark"
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
