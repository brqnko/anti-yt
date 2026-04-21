import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { lazy } from "preact-iso";
import { useTitle } from "../../hooks/useTitle";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { useRequireAuth } from "../../hooks/useRequireAuth";
import { useAuth } from "../../contexts/AuthContext";
import { DashboardLayout } from "../../components/DashboardLayout";
import { AuthPromptDialog } from "../../components/AuthPromptDialog";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { VideoCard } from "../../components/VideoCard";
import { getFeed } from "../../api/generated/feed";
import { getChannel } from "../../api/generated/channel";
import { getHistory } from "../../api/generated/history";
import { getPlaylist } from "../../api/generated/playlist";
import { PAGE_SIZES } from "../../constants";
import type { GetFeed200ItemsItem, GetPlaylistsRecent200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";
import { AddPlaylistDialog } from "../../components/AddPlaylistDialog";

const ScreenTimeBlock = lazy(() => import("../ScreenTimeBlock/index"));

function DashboardContent() {
  const { t } = useTranslation();
  useTitle(t("dashboard.pageTitle"));
  const { isAuthenticated, isLoading: isAuthLoading, requireAuth, showAuthPrompt, closeAuthPrompt } = useRequireAuth();
  const { screenTimeBlocked, screenTimeBlockReason } = useAuth();

  const [feedVideos, setFeedVideos] = useState<GetFeed200ItemsItem[]>([]);
  const [recentPlaylists, setRecentPlaylists] = useState<GetPlaylistsRecent200ItemsItem[]>([]);
  const [isLoadingFeed, setIsLoadingFeed] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [subscribedChannelIds, setSubscribedChannelIds] = useState<Set<string>>(new Set());
  const [showAddPlaylist, setShowAddPlaylist] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    if (isAuthLoading) return;
    const loadData = async () => {
      try {
        const [feedRes, subsRes, recentRes] = await Promise.all([
          getFeed().getFeed({ limit: PAGE_SIZES.FEED }).catch(() => null),
          isAuthenticated
            ? getChannel().getChannelsSubscribed({ limit: 50 }).catch(() => null)
            : Promise.resolve(null),
          isAuthenticated
            ? getPlaylist().getPlaylistsRecent().catch(() => null)
            : Promise.resolve(null),
        ]);
        if (feedRes) {
          setFeedVideos(feedRes.items);
          setHasNext(feedRes.has_next);
          const lastItem = feedRes.items[feedRes.items.length - 1];
          cursorRef.current = lastItem?.video_id;
        }
        if (subsRes) {
          setSubscribedChannelIds(new Set(subsRes.items.map((s) => s.channel_id)));
        }
        if (recentRes) {
          setRecentPlaylists(recentRes.items);
        }
      } finally {
        setIsLoadingFeed(false);
      }
    };
    loadData();
  }, [isAuthenticated, isAuthLoading]);

  const handleMarkWatched = useCallback(async (videoId: string) => {
    await getHistory().postVideosVideoIdWatched(videoId);
    setFeedVideos((prev) => prev.filter((v) => v.video_id !== videoId));
  }, []);

  const handleToggleSubscription = useCallback(async (channelId: string) => {
    const isCurrentlySubscribed = subscribedChannelIds.has(channelId);
    if (isCurrentlySubscribed) {
      await getChannel().deleteChannelsChannelIdSubscribe(channelId);
      setSubscribedChannelIds((prev) => {
        const next = new Set(prev);
        next.delete(channelId);
        return next;
      });
    } else {
      await getChannel().postChannelsSubscribe({ channel_id: channelId });
      setSubscribedChannelIds((prev) => new Set(prev).add(channelId));
    }
  }, [subscribedChannelIds]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasNext) return;
    setIsLoadingMore(true);
    try {
      const res = await getFeed().getFeed({ limit: PAGE_SIZES.FEED, cursor: cursorRef.current });
      setFeedVideos((prev) => [...prev, ...res.items]);
      setHasNext(res.has_next);
      const lastItem = res.items[res.items.length - 1];
      cursorRef.current = lastItem?.video_id;
    } finally {
      setIsLoadingMore(false);
    }
  }, [isLoadingMore, hasNext]);

  const sentinelRef = useInfiniteScroll(loadMore);

  if (isAuthenticated && screenTimeBlocked) {
    return <ScreenTimeBlock reason={screenTimeBlockReason ?? "limit_exceeded"} />;
  }

  return (
    <DashboardLayout>
      <div class="p-6">
        {recentPlaylists.length > 0 && (
        <div class="mb-8">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-bold">
              {t("dashboard.recentPlaylists")}
            </h3>
            <a
              href="/playlists"
              class="text-sm font-medium text-primary hover:text-primary/80 transition-colors no-underline"
            >
              {t("dashboard.showAllPlaylists")}
            </a>
          </div>
          <div class="flex gap-6 overflow-x-auto pb-2">
            {recentPlaylists.map((pl) => (
              <a
                key={pl.playlist_id}
                href={`/playlists/${pl.playlist_id}`}
                class="group flex-shrink-0 w-72 bg-card-light dark:bg-card-dark rounded-xl border border-transparent hover:border-primary/20 transition-all duration-300 overflow-hidden no-underline"
              >
                <div class="relative aspect-video w-full overflow-hidden bg-gray-100 dark:bg-gray-800">
                  {pl.top_video_thumbnail_url ? (
                    <img
                      src={pl.top_video_thumbnail_url}
                      alt={pl.playlist_title}
                      loading="lazy"
                      class="absolute inset-0 w-full h-full object-cover"
                    />
                  ) : (
                    <div class="absolute inset-0 flex items-center justify-center">
                      <Icon name="playlist_play" class="text-4xl text-text-muted-light dark:text-text-muted-dark" />
                    </div>
                  )}
                </div>
                <div class="p-3">
                  <h4 class="text-sm font-bold text-charcoal dark:text-white leading-tight line-clamp-2 group-hover:text-primary transition-colors">
                    {pl.playlist_title}
                  </h4>
                  <span class="text-xs text-text-muted-light dark:text-text-muted-dark mt-1 block">
                    {t("playlists.videoCount", { count: pl.playlist_video_count })}
                  </span>
                </div>
              </a>
            ))}
            {/* Create new playlist card */}
            <button
              type="button"
              class="group flex-shrink-0 w-72 rounded-xl border border-dashed border-border-light dark:border-border-dark hover:border-primary hover:bg-primary/5 transition-all duration-300 overflow-hidden bg-transparent cursor-pointer flex flex-col items-center justify-center gap-3"
              onClick={() => requireAuth(() => setShowAddPlaylist(true))}
            >
              <Icon name="add" class="text-4xl text-text-muted-light dark:text-text-muted-dark group-hover:text-primary transition-all" />
              <span class="text-sm font-bold text-text-muted-light dark:text-text-muted-dark group-hover:text-primary transition-colors">
                {t("dashboard.addPlaylist")}
              </span>
            </button>
          </div>
        </div>
        )}
        {isLoadingFeed ? (
          <LoadingSpinner />
        ) : feedVideos.length > 0 ? (
          <>
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
              {feedVideos.map((video) => (
                <VideoCard
                  key={video.video_id}
                  videoId={video.video_id}
                  thumbnailUrl={video.external_video_thumbnail_url}
                  title={video.external_video_title}
                  lengthSeconds={video.external_video_length_seconds}
                  channel={{
                    channelId: video.channel_id,
                    iconUrl: video.external_channel_icon_url,
                    displayName: video.external_channel_display_name,
                  }}
                  dateStr={video.external_video_created_at}
                  watchedSeconds={video.last_watch_seconds}
                  isSubscribed={subscribedChannelIds.has(video.channel_id)}
                  onToggleSubscription={() => requireAuth(() => handleToggleSubscription(video.channel_id))}
                  onMarkWatched={() => requireAuth(() => handleMarkWatched(video.video_id))}
                />
              ))}
            </div>
            <div ref={sentinelRef} class="h-1" />
            {isLoadingMore && <LoadingSpinner size="sm" className="py-8" />}
            {!hasNext && !isLoadingMore && (
              <p class="text-center text-sm text-text-muted-light dark:text-text-muted-dark py-8">
                🎉 {t("dashboard.endOfFeed")}
              </p>
            )}
          </>
        ) : (
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="subscriptions" class="text-5xl mb-4" />
            <p class="text-lg font-medium">{t("dashboard.noVideos")}</p>
            <p class="text-sm mt-1">{t("dashboard.noVideosDesc")}</p>
            <a
              class="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 transition-colors no-underline"
              href="/channels/explore"
            >
              <Icon name="recommend" class="text-[18px]" />
              {t("dashboard.nav.recommendedChannels")}
            </a>
          </div>
        )}
      </div>
      <AddPlaylistDialog
        open={showAddPlaylist}
        onClose={() => setShowAddPlaylist(false)}
        onAdded={async () => {
          const res = await getPlaylist().getPlaylistsRecent();
          setRecentPlaylists(res.items);
        }}
      />
      <AuthPromptDialog open={showAuthPrompt} onClose={closeAuthPrompt} />
    </DashboardLayout>
  );
}

export default function Dashboard() {
  return <DashboardContent />;
}
