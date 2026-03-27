import { useState, useEffect, useCallback } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import type { ComponentChildren } from "preact";
import useSWR from "swr";
import { DashboardHeader } from "./DashboardHeader";
import { AddChannelDialog } from "./AddChannelDialog";
import { AddPlaylistDialog } from "./AddPlaylistDialog";
import { getChannel } from "../api/generated/channel";
import { getPlaylist } from "../api/generated/playlist";
import { CACHE_KEYS } from "../api/cache-keys";
import { Icon } from "./Icon";
import type {
  GetChannelsSubscribed200ItemsItem,
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
  const [showAddChannel, setShowAddChannel] = useState(false);
  const [showAddPlaylist, setShowAddPlaylist] = useState(false);

  const [sidebarOpen, setSidebarOpen] = useState(getStoredSidebarState);

  const {
    data: subscriptions = [],
    isLoading: isSubscriptionsLoading,
    mutate: mutateSubscriptions,
  } = useSWR<GetChannelsSubscribed200ItemsItem[]>(
    CACHE_KEYS.dashboardSubscriptions,
    async () => {
      const res = await getChannel().getChannelsSubscribed({ limit: 10 });
      return res.items;
    },
  );

  const {
    data: playlists = [],
    isLoading: isPlaylistsLoading,
    mutate: mutatePlaylists,
  } = useSWR<GetPlaylists200ItemsItem[]>(
    CACHE_KEYS.dashboardPlaylists,
    async () => {
      const res = await getPlaylist().getPlaylists({ limit: 10 });
      return res.items;
    },
  );

  const isLoaded = !isSubscriptionsLoading && !isPlaylistsLoading;

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
                  url === "/dashboard"
                    ? "bg-primary/10 text-primary"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href="/dashboard"
                aria-current={url === "/dashboard" ? "page" : undefined}
              >
                <Icon name="home" />
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
                <Icon name="analytics" />
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
                <Icon name="recommend" />
                {t("dashboard.nav.recommendedChannels")}
              </a>
              <a
                class={`flex items-center gap-3 px-3 py-2 rounded-lg no-underline ${
                  url === "/playlists" || url.startsWith("/playlists/")
                    ? "bg-primary/10 text-primary font-bold"
                    : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white font-medium transition-colors"
                }`}
                href="/playlists"
                aria-current={url === "/playlists" || url.startsWith("/playlists/") ? "page" : undefined}
              >
                <Icon name="playlist_play" />
                {t("dashboard.nav.playlists")}
              </a>

            </nav>

            <div class="h-px bg-border-light dark:bg-border-dark w-full my-2" role="separator" />

            {/* Whitelisted Channels */}
            <div class="flex flex-col gap-4">
              <h3 class="text-xs font-bold uppercase tracking-wider text-text-muted-light dark:text-text-muted-dark px-3">
                {t("dashboard.whitelistedChannels")}
              </h3>
              <div class="flex flex-col gap-2 max-h-32 overflow-y-auto">
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
                        loading="lazy"
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
              </div>
              <button
                  class="flex items-center gap-2 px-3 py-2 text-sm text-primary font-medium mt-1 cursor-pointer bg-transparent border-none group/add"
                  onClick={() => setShowAddChannel(true)}
                >
                  <Icon name="add" class="text-[18px]" />
                  <span class="group-hover/add:underline">{t("dashboard.requestChannel")}</span>
                </button>
            </div>

            <div class="h-px bg-border-light dark:bg-border-dark w-full my-2" role="separator" />

            {/* Playlists */}
            <div class="flex flex-col gap-4">
              <h3 class="text-xs font-bold uppercase tracking-wider text-text-muted-light dark:text-text-muted-dark px-3">
                {t("dashboard.myPlaylists")}
              </h3>
              <div class="flex flex-col gap-2 max-h-36 overflow-y-auto">
                {playlists.map((pl) => (
                  <a
                    key={pl.playlist_id}
                    class={`flex items-center gap-3 px-3 py-2 hover:bg-card-light dark:hover:bg-card-dark rounded-lg group transition-colors no-underline ${
                      url === `/playlists/${pl.playlist_id}` ? "bg-primary/10" : ""
                    }`}
                    href={`/playlists/${pl.playlist_id}`}
                  >
                    <Icon name="playlist_play" class="text-text-muted-light dark:text-text-muted-dark group-hover:text-primary" />
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
              </div>
              <button
                class="flex items-center gap-2 px-3 py-2 text-sm text-primary font-medium mt-1 cursor-pointer bg-transparent border-none group/add"
                onClick={() => setShowAddPlaylist(true)}
              >
                <Icon name="add" class="text-[18px]" />
                <span class="group-hover/add:underline">{t("dashboard.addPlaylist")}</span>
              </button>
            </div>
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
          href="/dashboard"
          onClick={(e) => {
            if (url === "/dashboard") {
              e.preventDefault();
              window.scrollTo({ top: 0, behavior: "smooth" });
            }
          }}
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors ${
            url === "/dashboard"
              ? "text-primary font-bold"
              : "text-text-muted-light dark:text-text-muted-dark"
          }`}
          aria-current={url === "/dashboard" ? "page" : undefined}
        >
          <Icon name="home" class="text-xl" />
          {t("dashboard.nav.bottomFeed")}
        </a>
        <a
          href="/channels"
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors ${
            url === "/channels"
              ? "text-primary font-bold"
              : "text-text-muted-light dark:text-text-muted-dark"
          }`}
          aria-current={url === "/channels" ? "page" : undefined}
        >
          <Icon name="recommend" class="text-xl" />
          {t("dashboard.nav.bottomChannels")}
        </a>
        <a
          href="/playlists"
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors ${
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
          href="/analytics"
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors ${
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
          href="/profile"
          onClick={(e) => {
            if (url === "/profile") {
              e.preventDefault();
              window.scrollTo({ top: 0, behavior: "smooth" });
            }
          }}
          class={`flex flex-col items-center gap-0.5 py-2 px-3 text-[10px] no-underline transition-colors ${
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

      <AddChannelDialog
        open={showAddChannel}
        onClose={() => setShowAddChannel(false)}
        onAdded={() => mutateSubscriptions()}
      />
      <AddPlaylistDialog
        open={showAddPlaylist}
        onClose={() => setShowAddPlaylist(false)}
        onAdded={() => mutatePlaylists()}
      />
    </div>
  );
}
