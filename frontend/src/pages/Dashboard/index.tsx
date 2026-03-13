import { useState, useEffect } from "preact/hooks";
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
} from "../../api/generated/antiYtApi.schemas";

function formatDuration(totalSeconds: number): string {
  const h = Math.floor(totalSeconds / 3600);
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = totalSeconds % 60;
  const mm = String(m).padStart(2, "0");
  const ss = String(s).padStart(2, "0");
  return h > 0 ? `${h}:${mm}:${ss}` : `${mm}:${ss}`;
}

function formatViews(count: number): string {
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1).replace(/\.0$/, "")}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1).replace(/\.0$/, "")}K`;
  return String(count);
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
    <article class="flex flex-col gap-3 group">
      <div class="relative aspect-video rounded-xl overflow-hidden bg-gray-200 dark:bg-gray-800">
        <div
          class="absolute inset-0 bg-cover bg-center transition-transform duration-500 group-hover:scale-105"
          style={`background-image: url('${video.external_video_thumbnail_url}');`}
        />
        <div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
        <span class="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-bold px-1.5 py-0.5 rounded">
          {formatDuration(video.external_video_length_seconds)}
        </span>
        <div class="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none">
          <div class="size-12 rounded-full bg-primary/90 flex items-center justify-center text-white shadow-lg transform scale-90 group-hover:scale-100 transition-transform">
            <span class="material-symbols-outlined text-[28px] ml-1">
              play_arrow
            </span>
          </div>
        </div>
      </div>
      <div class="flex gap-3">
        <div class="size-9 rounded-full bg-gray-300 dark:bg-gray-700 flex-shrink-0 overflow-hidden mt-1">
          <img
            alt={video.external_channel_display_name}
            class="w-full h-full object-cover"
            src={video.external_channel_icon_url}
          />
        </div>
        <div class="flex flex-col">
          <h3 class="text-base font-bold text-charcoal dark:text-white leading-tight line-clamp-2 group-hover:text-primary transition-colors">
            {video.external_video_title}
          </h3>
          <div class="flex flex-col text-sm text-text-muted-light dark:text-text-muted-dark mt-1">
            <span class="font-medium hover:text-charcoal dark:hover:text-white cursor-pointer">
              {video.external_channel_display_name}
            </span>
            <div class="flex items-center gap-1 text-xs mt-0.5">
              <span>
                {formatViews(video.external_video_watch_count)}{" "}
                {t("dashboard.views")}
              </span>
              <span>-</span>
              <span>{formatTimeAgo(video.external_video_created_at, t)}</span>
            </div>
          </div>
        </div>
      </div>
    </article>
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

  useEffect(() => {
    const loadData = async () => {
      try {
        const [feedRes, subsRes, playlistRes] = await Promise.allSettled([
          getVideo().getFeed({ limit: 12 }),
          getChannel().getSubscriptions({ limit: 10 }),
          getPlaylist().getPlaylists({ limit: 10 }),
        ]);
        if (feedRes.status === "fulfilled") setFeedVideos(feedRes.value.items);
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

  return (
    <div class="relative flex min-h-screen w-full flex-col overflow-x-hidden bg-background-light dark:bg-background-dark text-charcoal dark:text-white font-display antialiased">
      <DashboardHeader />

      <div class="flex flex-1 w-full max-w-[1600px] mx-auto">
        {/* Left Sidebar */}
        <aside class="hidden lg:flex w-64 flex-col gap-6 p-6 border-r border-border-light dark:border-border-dark sticky top-[65px] h-[calc(100vh-65px)] overflow-y-auto">
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
              href="#"
            >
              <span class="material-symbols-outlined">settings</span>
              {t("dashboard.nav.settings")}
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
                  href="#"
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
              <button class="flex items-center gap-2 px-3 py-2 text-sm text-primary font-medium hover:underline mt-1 cursor-pointer bg-transparent border-none">
                <span class="material-symbols-outlined text-[18px]">add</span>
                {t("dashboard.requestChannel")}
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
        <main class="flex-1 flex flex-col p-6 min-w-0">

          {/* Feed header */}
          <div class="mb-6">
            <h1 class="text-2xl font-bold">{t("dashboard.curatedFeed")}</h1>
          </div>

          {/* Video grid */}
          {isLoadingFeed ? (
            <div class="flex items-center justify-center py-20">
              <span class="material-symbols-outlined text-5xl animate-spin text-primary">
                progress_activity
              </span>
            </div>
          ) : feedVideos.length > 0 ? (
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
              {feedVideos.map((video) => (
                <VideoCard key={video.video_id} video={video} t={t} />
              ))}
            </div>
          ) : (
            <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
              <span class="material-symbols-outlined text-5xl mb-4">
                subscriptions
              </span>
              <p class="text-lg font-medium">{t("dashboard.noVideos")}</p>
              <p class="text-sm mt-1">{t("dashboard.noVideosDesc")}</p>
            </div>
          )}
        </main>

      </div>
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
