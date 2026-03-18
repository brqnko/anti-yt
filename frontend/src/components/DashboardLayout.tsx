import { useState, useEffect, useCallback } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import type { ComponentChildren } from "preact";
import { DashboardHeader } from "./DashboardHeader";
import { AddChannelDialog } from "./AddChannelDialog";
import { AddPlaylistDialog } from "./AddPlaylistDialog";
import { getChannel } from "../api/generated/channel";
import { getPlaylist } from "../api/generated/playlist";
import type {
  GetSubscriptions200ItemsItem,
  GetPlaylists200ItemsItem,
} from "../api/generated/antiYtApi.schemas";

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
  const [subscriptions, setSubscriptions] = useState<GetSubscriptions200ItemsItem[]>([]);
  const [playlists, setPlaylists] = useState<GetPlaylists200ItemsItem[]>([]);
  const [isLoaded, setIsLoaded] = useState(false);
  const [showAddChannel, setShowAddChannel] = useState(false);
  const [showAddPlaylist, setShowAddPlaylist] = useState(false);

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

  useEffect(() => {
    const load = async () => {
      try {
        const [subsRes, playlistRes] = await Promise.allSettled([
          getChannel().getSubscriptions({ limit: 10 }),
          getPlaylist().getPlaylists({ limit: 10 }),
        ]);
        if (subsRes.status === "fulfilled") setSubscriptions(subsRes.value.items);
        if (playlistRes.status === "fulfilled") setPlaylists(playlistRes.value.items);
      } finally {
        setIsLoaded(true);
      }
    };
    load();
  }, []);

  return (
    <div class="relative flex h-screen w-full flex-col overflow-hidden bg-background-light dark:bg-background-dark text-charcoal dark:text-white font-display antialiased">
      <DashboardHeader sidebarOpen={sidebarOpen} onToggleSidebar={toggleSidebar} />

      <div class="flex flex-1 w-full max-w-[1600px] mx-auto overflow-hidden">
        {/* Mobile backdrop */}
        {sidebarOpen && (
          <div
            class="fixed inset-0 z-30 bg-black/40 lg:hidden"
            onClick={toggleSidebar}
          />
        )}

        {/* Sidebar */}
        <aside
          class={`flex flex-col border-r border-border-light dark:border-border-dark shrink-0 transition-[width,opacity,transform] duration-200
            fixed top-[57px] bottom-0 z-40 bg-background-light dark:bg-background-dark
            lg:relative lg:top-auto lg:bottom-auto lg:z-auto
            ${sidebarOpen
              ? "w-64 opacity-100 translate-x-0 overflow-y-auto"
              : "w-0 opacity-0 -translate-x-full lg:translate-x-0 overflow-hidden"
            }`}
          role="navigation"
          aria-label={t("dashboard.nav.sidebar")}
        >
          <div class="flex flex-col gap-6 px-6 pb-6 pt-4 min-w-[16rem]">
            <nav class="flex flex-col gap-1" aria-label={t("dashboard.nav.mainNav")}>
              <a
                class={`flex items-center gap-3 px-3 py-2 rounded-lg font-bold no-underline ${
                  url === "/dashboard"
                    ? "bg-primary/10 text-primary"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href="/dashboard"
                aria-current={url === "/dashboard" ? "page" : undefined}
              >
                <span class="material-symbols-outlined">home</span>
                {t("dashboard.nav.mainFeed")}
              </a>
              <a
                class={`flex items-center gap-3 px-3 py-2 rounded-lg no-underline ${
                  url === "/analytics"
                    ? "bg-primary/10 text-primary font-bold"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href="/analytics"
                aria-current={url === "/analytics" ? "page" : undefined}
              >
                <span class="material-symbols-outlined">analytics</span>
                {t("dashboard.nav.analytics")}
              </a>
              <a
                class={`flex items-center gap-3 px-3 py-2 rounded-lg no-underline ${
                  url === "/channels"
                    ? "bg-primary/10 text-primary font-bold"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href="/channels"
                aria-current={url === "/channels" ? "page" : undefined}
              >
                <span class="material-symbols-outlined">recommend</span>
                {t("dashboard.nav.recommendedChannels")}
              </a>
              <a
                class={`flex items-center gap-3 px-3 py-2 rounded-lg no-underline ${
                  url === "/history"
                    ? "bg-primary/10 text-primary font-bold"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href="/history"
                aria-current={url === "/history" ? "page" : undefined}
              >
                <span class="material-symbols-outlined">history</span>
                {t("dashboard.nav.history")}
              </a>
            </nav>

            <div class="h-px bg-border-light dark:bg-border-dark w-full my-2" role="separator" />

            {/* Whitelisted Channels */}
            <div class="flex flex-col gap-4">
              <h3 class="text-xs font-bold uppercase tracking-wider text-text-muted-light dark:text-text-muted-dark px-3">
                {t("dashboard.whitelistedChannels")}
              </h3>
              <div class="flex flex-col gap-2">
                {subscriptions.map((sub) => (
                  <a
                    key={sub.channel_id}
                    class={`flex items-center gap-3 px-3 py-2 hover:bg-card-light dark:hover:bg-card-dark rounded-lg group transition-colors no-underline ${
                      url === `/channels/${sub.channel_id}` ? "bg-primary/10" : ""
                    }`}
                    href={`/channels/${sub.channel_id}`}
                    aria-current={url === `/channels/${sub.channel_id}` ? "page" : undefined}
                  >
                    <div class="size-8 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden shrink-0">
                      <img
                        alt={sub.external_channel_display_name}
                        class="w-full h-full object-cover"
                        src={sub.external_channel_icon_url}
                      />
                    </div>
                    <span class="text-sm font-medium text-charcoal dark:text-white group-hover:text-primary transition-colors truncate">
                      {sub.external_channel_display_name}
                    </span>
                  </a>
                ))}
                {subscriptions.length === 0 && isLoaded && (
                  <p class="text-xs text-text-muted-light dark:text-text-muted-dark px-3">
                    {t("dashboard.noChannels")}
                  </p>
                )}
                <button
                  class="flex items-center gap-2 px-3 py-2 text-sm text-primary font-medium mt-1 cursor-pointer bg-transparent border-none group/add"
                  onClick={() => setShowAddChannel(true)}
                >
                  <span class="material-symbols-outlined text-[18px]">add</span>
                  <span class="group-hover/add:underline">{t("dashboard.requestChannel")}</span>
                </button>
              </div>
            </div>

            <div class="h-px bg-border-light dark:bg-border-dark w-full my-2" role="separator" />

            {/* Playlists */}
            <div class="flex flex-col gap-4">
              <h3 class="text-xs font-bold uppercase tracking-wider text-text-muted-light dark:text-text-muted-dark px-3">
                {t("dashboard.myPlaylists")}
              </h3>
              <div class="flex flex-col gap-2">
                {playlists.map((pl) => (
                  <a
                    key={pl.playlist_id}
                    class="flex items-center gap-3 px-3 py-2 hover:bg-card-light dark:hover:bg-card-dark rounded-lg group transition-colors no-underline"
                    href="#"
                  >
                    <span class="material-symbols-outlined text-text-muted-light dark:text-text-muted-dark group-hover:text-primary">
                      playlist_play
                    </span>
                    <span class="text-sm font-medium text-charcoal dark:text-white truncate">
                      {pl.playlist_title}
                    </span>
                  </a>
                ))}
                {playlists.length === 0 && isLoaded && (
                  <p class="text-xs text-text-muted-light dark:text-text-muted-dark px-3">
                    {t("dashboard.noPlaylists")}
                  </p>
                )}
                <button
                  class="flex items-center gap-2 px-3 py-2 text-sm text-primary font-medium mt-1 cursor-pointer bg-transparent border-none group/add"
                  onClick={() => setShowAddPlaylist(true)}
                >
                  <span class="material-symbols-outlined text-[18px]">add</span>
                  <span class="group-hover/add:underline">{t("dashboard.addPlaylist")}</span>
                </button>
              </div>
            </div>
          </div>
        </aside>

        {/* Main Content */}
        <main class="flex-1 flex flex-col min-w-0 overflow-y-auto">
          {children}
        </main>
      </div>

      <AddChannelDialog
        open={showAddChannel}
        onClose={() => setShowAddChannel(false)}
        onAdded={(sub) => setSubscriptions((prev) => [...prev, sub])}
      />
      <AddPlaylistDialog
        open={showAddPlaylist}
        onClose={() => setShowAddPlaylist(false)}
        onAdded={(pl) => setPlaylists((prev) => [...prev, pl])}
      />
    </div>
  );
}
