import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardHeader } from "../../components/DashboardHeader";
import { getVideo } from "../../api/generated/video";
import { getChannel } from "../../api/generated/channel";
import { getPlaylist } from "../../api/generated/playlist";
import type {
  GetFeed200ItemsItem,
  GetSubscriptions200ItemsItem,
  GetPlaylists200ItemsItem,
  PostSubscriptions201,
} from "../../api/generated/antiYtApi.schemas";

function formatDuration(totalSeconds: number): string {
  const h = Math.floor(totalSeconds / 3600);
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = totalSeconds % 60;
  const mm = String(m).padStart(2, "0");
  const ss = String(s).padStart(2, "0");
  return h > 0 ? `${h}:${mm}:${ss}` : `${mm}:${ss}`;
}

function formatTimeAgo(dateStr: string, t: (key: string, opts?: object) => string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);
  const weeks = Math.floor(days / 7);
  const months = Math.floor(days / 30);
  const years = Math.floor(days / 365);

  if (years > 0) return t("dashboard.timeAgo.years", { count: years });
  if (months > 0) return t("dashboard.timeAgo.months", { count: months });
  if (weeks > 0) return t("dashboard.timeAgo.weeks", { count: weeks });
  if (days > 0) return t("dashboard.timeAgo.days", { count: days });
  if (hours > 0) return t("dashboard.timeAgo.hours", { count: hours });
  if (minutes > 0) return t("dashboard.timeAgo.minutes", { count: minutes });
  return t("dashboard.timeAgo.justNow");
}

function VideoCard({
  video,
  t,
}: {
  video: GetFeed200ItemsItem;
  t: (key: string, opts?: object) => string;
}) {
  return (
    <article class="flex flex-col gap-3">
      <div class="group/thumb relative aspect-video rounded-xl overflow-hidden bg-gray-200 dark:bg-gray-800 cursor-pointer">
        <img
          src={video.external_video_thumbnail_url}
          alt={video.external_video_title}
          class="absolute inset-0 w-full h-full object-cover transition-transform duration-500 group-hover/thumb:scale-105"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover/thumb:opacity-100 transition-opacity duration-300" />
        <span class="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-bold px-1.5 py-0.5 rounded">
          {formatDuration(video.external_video_length_seconds)}
        </span>
        <div class="absolute inset-0 flex items-center justify-center opacity-0 group-hover/thumb:opacity-100 transition-opacity duration-300 pointer-events-none">
          <div class="size-12 rounded-full bg-primary/90 flex items-center justify-center text-white shadow-lg transform scale-90 group-hover/thumb:scale-100 transition-transform">
            <span class="material-symbols-outlined text-[28px] ml-1">
              play_arrow
            </span>
          </div>
        </div>
      </div>
      <div class="flex gap-3 items-start">
        <a href={`/channels/${video.channel_id}`} class="size-9 rounded-full bg-gray-300 dark:bg-gray-700 flex-shrink-0 overflow-hidden cursor-pointer">
          <img
            alt={video.external_channel_display_name}
            class="w-full h-full object-cover"
            src={video.external_channel_icon_url}
          />
        </a>
        <div class="flex flex-col min-w-0 flex-1">
          <h3 class="text-base font-bold text-charcoal dark:text-white leading-tight line-clamp-2 cursor-pointer">
            {video.external_video_title}
          </h3>
          <div class="flex items-center justify-between text-sm text-text-muted-light dark:text-text-muted-dark mt-1">
            <a href={`/channels/${video.channel_id}`} class="font-medium hover:text-charcoal dark:hover:text-white cursor-pointer truncate no-underline text-text-muted-light dark:text-text-muted-dark">
              {video.external_channel_display_name}
            </a>
            <span class="text-xs flex-shrink-0 ml-2">{formatTimeAgo(video.external_video_created_at, t)}</span>
          </div>
        </div>
      </div>
    </article>
  );
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
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} />
      <div class="relative bg-white dark:bg-[#2a2721] rounded-2xl shadow-2xl border border-gray-100 dark:border-neutral-800 p-8 max-w-sm w-full">
        <button
          class="absolute top-4 right-4 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors bg-transparent border-none cursor-pointer"
          onClick={onClose}
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
          <p class="text-sm text-red-500 mt-2">{error}</p>
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

function DashboardContent() {
  const { t } = useTranslation();
  useTitle(t("dashboard.pageTitle"));

  const [feedVideos, setFeedVideos] = useState<GetFeed200ItemsItem[]>([]);
  const [subscriptions, setSubscriptions] = useState<
    GetSubscriptions200ItemsItem[]
  >([]);
  const [playlists, setPlaylists] = useState<GetPlaylists200ItemsItem[]>([]);
  const [isLoadingFeed, setIsLoadingFeed] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);
  const [showAddChannel, setShowAddChannel] = useState(false);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [feedRes, subsRes, playlistRes] = await Promise.allSettled([
          getVideo().getFeed({ limit: 12 }),
          getChannel().getSubscriptions({ limit: 10 }),
          getPlaylist().getPlaylists({ limit: 10 }),
        ]);
        if (feedRes.status === "fulfilled") {
          setFeedVideos(feedRes.value.items);
          setHasNext(feedRes.value.has_next);
          const lastItem = feedRes.value.items[feedRes.value.items.length - 1];
          cursorRef.current = lastItem?.video_id;
        }
        if (subsRes.status === "fulfilled")
          setSubscriptions(subsRes.value.items);
        if (playlistRes.status === "fulfilled")
          setPlaylists(playlistRes.value.items);
      } finally {
        setIsLoadingFeed(false);
      }
    };
    loadData();
  }, []);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasNext) return;
    setIsLoadingMore(true);
    try {
      const res = await getVideo().getFeed({ limit: 12, cursor: cursorRef.current });
      setFeedVideos((prev) => [...prev, ...res.items]);
      setHasNext(res.has_next);
      const lastItem = res.items[res.items.length - 1];
      cursorRef.current = lastItem?.video_id;
    } finally {
      setIsLoadingMore(false);
    }
  }, [isLoadingMore, hasNext]);

  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) loadMore();
      },
      { rootMargin: "200px" },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [loadMore]);

  return (
    <div class="relative flex h-screen w-full flex-col overflow-hidden bg-background-light dark:bg-background-dark text-charcoal dark:text-white font-display antialiased">
      <DashboardHeader />

      <div class="flex flex-1 w-full max-w-[1600px] mx-auto overflow-hidden">
        {/* Left Sidebar */}
        <aside class="hidden lg:flex w-64 flex-col gap-6 p-6 border-r border-border-light dark:border-border-dark overflow-y-auto shrink-0">
          <nav class="flex flex-col gap-1">
            <a
              class="flex items-center gap-3 px-3 py-2 bg-primary/10 text-primary rounded-lg font-bold no-underline"
              href="/dashboard"
            >
              <span class="material-symbols-outlined">home</span>
              {t("dashboard.nav.mainFeed")}
            </a>
            <a
              class="flex items-center gap-3 px-3 py-2 text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white rounded-lg font-medium transition-colors no-underline"
              href="#"
            >
              <span class="material-symbols-outlined">analytics</span>
              {t("dashboard.nav.analytics")}
            </a>
            <a
              class="flex items-center gap-3 px-3 py-2 text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 hover:text-charcoal dark:hover:text-white rounded-lg font-medium transition-colors no-underline"
              href="/channels"
            >
              <span class="material-symbols-outlined">recommend</span>
              {t("dashboard.nav.recommendedChannels")}
            </a>

          </nav>

          <div class="h-px bg-border-light dark:bg-border-dark w-full my-2" />

          {/* Whitelisted Channels */}
          <div class="flex flex-col gap-4">
            <h3 class="text-xs font-bold uppercase tracking-wider text-text-muted-light dark:text-text-muted-dark px-3">
              {t("dashboard.whitelistedChannels")}
            </h3>
            <div class="flex flex-col gap-2">
              {subscriptions.map((sub) => (
                <a
                  key={sub.channel_id}
                  class="flex items-center gap-3 px-3 py-2 hover:bg-card-light dark:hover:bg-card-dark rounded-lg group transition-colors no-underline"
                  href={`/channels/${sub.channel_id}`}
                >
                  <div class="size-8 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden">
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
              {subscriptions.length === 0 && !isLoadingFeed && (
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

          <div class="h-px bg-border-light dark:bg-border-dark w-full my-2" />

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
              {playlists.length === 0 && !isLoadingFeed && (
                <p class="text-xs text-text-muted-light dark:text-text-muted-dark px-3">
                  {t("dashboard.noPlaylists")}
                </p>
              )}
            </div>
          </div>
        </aside>

        {/* Main Content */}
        <main class="flex-1 flex flex-col p-6 min-w-0 overflow-y-auto">

          {/* Video grid */}
          {isLoadingFeed ? (
            <div class="flex items-center justify-center py-20">
              <span class="material-symbols-outlined text-5xl animate-spin text-primary">
                progress_activity
              </span>
            </div>
          ) : feedVideos.length > 0 ? (
            <>
              <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
                {feedVideos.map((video) => (
                  <VideoCard key={video.video_id} video={video} t={t} />
                ))}
              </div>
              <div ref={sentinelRef} class="h-1" />
              {isLoadingMore && (
                <div class="flex items-center justify-center py-8">
                  <span class="material-symbols-outlined text-3xl animate-spin text-primary">
                    progress_activity
                  </span>
                </div>
              )}
              {!hasNext && !isLoadingMore && (
                <p class="text-center text-sm text-text-muted-light dark:text-text-muted-dark py-8">
                  🎉 {t("dashboard.endOfFeed")}
                </p>
              )}
            </>
          ) : (
            <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
              <span class="material-symbols-outlined text-5xl mb-4">
                subscriptions
              </span>
              <p class="text-lg font-medium">{t("dashboard.noVideos")}</p>
              <p class="text-sm mt-1">{t("dashboard.noVideosDesc")}</p>
              <a
                class="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 transition-colors no-underline"
                href="/channels"
              >
                <span class="material-symbols-outlined text-[18px]">recommend</span>
                {t("dashboard.nav.recommendedChannels")}
              </a>
            </div>
          )}
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

export default function Dashboard() {
  return (
    <ProtectedRoute>
      <DashboardContent />
    </ProtectedRoute>
  );
}
