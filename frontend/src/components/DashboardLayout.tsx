import { useState, useEffect, useCallback } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import type { ComponentChildren } from "preact";
import { DashboardHeader } from "./DashboardHeader";
import { getChannel } from "../api/generated/channel";
import { getPlaylist } from "../api/generated/playlist";
import type {
  GetSubscriptions200ItemsItem,
  GetPlaylists200ItemsItem,
  PostSubscriptions201,
} from "../api/generated/antiYtApi.schemas";

const SIDEBAR_STORAGE_KEY = "sidebar-open";

function getStoredSidebarState(): boolean {
  try {
    const stored = localStorage.getItem(SIDEBAR_STORAGE_KEY);
    if (stored !== null) return stored === "true";
  } catch {}
  return true;
}

function AddChannelDialog({
  open,
  onClose,
  onAdded,
}: {
  open: boolean;
  onClose: () => void;
  onAdded: (sub: PostSubscriptions201) => void;
}) {
  const { t } = useTranslation();
  const [channelId, setChannelId] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) {
      setChannelId("");
      setIsSubmitting(false);
      setError(null);
    }
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, onClose]);

  if (!open) return null;

  const handleSubmit = async () => {
    const trimmed = channelId.trim();
    if (!trimmed || isSubmitting) return;
    setIsSubmitting(true);
    setError(null);
    try {
      const result = await getChannel().postSubscriptions({ channel_id: trimmed });
      onAdded(result);
      onClose();
    } catch {
      setError(t("dashboard.addChannelDialog.error"));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4" role="dialog" aria-modal="true" aria-label={t("dashboard.addChannelDialog.title")}>
      <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} />
      <div class="relative bg-white dark:bg-[#2a2721] rounded-2xl shadow-2xl border border-gray-100 dark:border-neutral-800 p-8 max-w-sm w-full">
        <button
          class="absolute top-4 right-4 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors bg-transparent border-none cursor-pointer"
          onClick={onClose}
          aria-label={t("dashboard.addChannelDialog.cancel")}
        >
          <span class="material-symbols-outlined">close</span>
        </button>
        <h2 class="text-lg font-bold text-charcoal dark:text-white mb-2">
          {t("dashboard.addChannelDialog.title")}
        </h2>
        <p class="text-sm text-text-muted-light dark:text-text-muted-dark mb-4">
          {t("dashboard.addChannelDialog.description")}
        </p>
        <div class="relative">
          <button
            type="button"
            class="absolute inset-y-0 left-0 flex items-center pl-3 pr-1 text-text-muted-light dark:text-text-muted-dark hover:text-primary transition-colors bg-transparent border-none cursor-pointer"
            aria-label={t("dashboard.addChannelDialog.paste")}
            onClick={async () => {
              try {
                const text = await navigator.clipboard.readText();
                if (text) setChannelId(text);
              } catch {}
            }}
          >
            <span class="material-symbols-outlined text-[20px]">content_paste</span>
          </button>
          <input
            type="text"
            class="w-full pl-10 pr-4 py-3 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white placeholder-taupe focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all shadow-sm"
            placeholder={t("dashboard.addChannelDialog.placeholder")}
            value={channelId}
            onInput={(e) => setChannelId((e.target as HTMLInputElement).value)}
            onKeyDown={(e) => { if (e.key === "Enter") handleSubmit(); }}
          />
        </div>
        {error && (
          <p class="text-sm text-red-500 mt-2" role="alert">{error}</p>
        )}
        <div class="flex justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 rounded-xl text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors bg-transparent border-none cursor-pointer"
            onClick={onClose}
          >
            {t("dashboard.addChannelDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-xl text-sm font-bold text-white bg-primary hover:bg-primary/90 transition-colors border-none cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={!channelId.trim() || isSubmitting}
            onClick={handleSubmit}
          >
            {isSubmitting
              ? t("dashboard.addChannelDialog.adding")
              : t("dashboard.addChannelDialog.add")}
          </button>
        </div>
      </div>
    </div>
  );
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
    </div>
  );
}
