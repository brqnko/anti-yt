import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { memo } from "preact/compat";
import { useTranslation } from "react-i18next";
import { useMeta } from "../../hooks/useMeta";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { useRequireAuth } from "../../hooks/useRequireAuth";
import { DashboardLayout } from "../../components/DashboardLayout";
import { AuthPromptDialog } from "../../components/AuthPromptDialog";
import { ChannelInfoCard } from "../../components/ChannelInfoCard";
import { getChannel } from "../../api/generated/channel";
import { getHistory } from "../../api/generated/history";
import { VideoCard } from "../../components/VideoCard";
import {
  GetChannelsChannelIdVideosOrder,
  type GetChannelsChannelIdVideosOrder as ChannelVideoOrder,
  GetChannelsChannelId200,
  GetChannelsChannelIdPlaylists200ItemsItem,
  GetChannelsChannelIdVideos200ItemsItem,
} from "../../api/generated/antiYtApi.schemas";
import { PAGE_SIZES } from "../../constants";
import { Icon } from "../../components/Icon";
import { BrowserBackLink } from "../../components/BrowserBackLink";
import {
  ChannelInfoCardSkeleton,
  ChannelDetailPlaylistCardSkeleton,
  VideoCardSkeleton,
  SkeletonRepeat,
} from "../../components/skeletons";

const CHANNEL_VIDEO_ORDER_OPTIONS = Object.values(GetChannelsChannelIdVideosOrder);

function parseChannelVideoOrder(value: string): ChannelVideoOrder | undefined {
  return CHANNEL_VIDEO_ORDER_OPTIONS.find((order) => order === value);
}

const ChannelVideoCard = memo(function ChannelVideoCard({
  video,
  isWatched,
  requireAuth,
  onToggleWatched,
}: {
  video: GetChannelsChannelIdVideos200ItemsItem;
  isWatched: boolean;
  requireAuth: (fn: () => void | Promise<void>) => Promise<void>;
  onToggleWatched: (videoId: string) => Promise<void>;
}) {
  const handleMark = useCallback(
    () => requireAuth(() => onToggleWatched(video.video_id)),
    [requireAuth, onToggleWatched, video.video_id],
  );
  return (
    <VideoCard
      videoId={video.video_id}
      thumbnailUrl={video.external_video_thumbnail_url}
      title={video.external_video_title}
      lengthSeconds={video.external_video_length_seconds}
      dateStr={video.external_video_created_at}
      watchedSeconds={video.last_watch_seconds}
      isWatched={isWatched}
      onMarkWatched={handleMark}
    />
  );
});

function ChannelDetailContent({ channelId }: { channelId: string }) {
  const { t } = useTranslation();
  const { isAuthenticated, isLoading: isAuthLoading, requireAuth, showAuthPrompt, closeAuthPrompt } = useRequireAuth();

  const [channelInfo, setChannelInfo] = useState<GetChannelsChannelId200 | null>(null);
  const [playlists, setPlaylists] = useState<GetChannelsChannelIdPlaylists200ItemsItem[]>([]);
  const [videos, setVideos] = useState<GetChannelsChannelIdVideos200ItemsItem[]>([]);
  const [isSubscribed, setIsSubscribed] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isVideosLoading, setIsVideosLoading] = useState(false);
  const [isToggling, setIsToggling] = useState(false);
  const [hasNextVideos, setHasNextVideos] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [order, setOrder] = useState<ChannelVideoOrder>("newer");
  const [watchedVideoIds, setWatchedVideoIds] = useState<Set<string>>(new Set());
  const cursorRef = useRef<string | undefined>(undefined);
  const channelUuidRef = useRef<string | undefined>(undefined);

  useMeta({
    title: channelInfo?.external_channel_display_name ?? t("channelDetail.pageTitle"),
    description: channelInfo?.external_channel_description?.slice(0, 160) || t("channelDetail.metaDescription"),
    canonicalPath: `/channels/${channelId}`,
    ogImage: channelInfo?.external_channel_icon_url,
  });

  useEffect(() => {
    setOrder("newer");
    setVideos([]);
    setHasNextVideos(false);
    cursorRef.current = undefined;
    channelUuidRef.current = undefined;
  }, [channelId]);

  useEffect(() => {
    if (isAuthLoading) return;
    const load = async () => {
      setIsVideosLoading(true);
      try {
        const channelRes = await getChannel().getChannelsChannelId(channelId).catch(() => null);
        if (!channelRes) return;
        setChannelInfo(channelRes);
        const uuid = channelRes.channel_id;
        channelUuidRef.current = uuid;

        const [videosRes, playlistsRes, subsRes] = await Promise.all([
          getChannel().getChannelsChannelIdVideos(uuid, { limit: PAGE_SIZES.CHANNEL_VIDEOS, order }).catch(() => null),
          getChannel().getChannelsChannelIdPlaylists(uuid, { limit: PAGE_SIZES.CHANNEL_PLAYLISTS }).catch(() => null),
          isAuthenticated
            ? getChannel().getChannelsSubscribed({ limit: 50 }).catch(() => null)
            : Promise.resolve(null),
        ]);

        if (subsRes) {
          const found = subsRes.items.find(s => s.channel_id === uuid);
          if (found) {
            setIsSubscribed(true);
          }
        }

        if (videosRes) {
          setVideos(videosRes.items);
          setHasNextVideos(videosRes.has_next);
          setWatchedVideoIds(new Set(videosRes.items.filter(v => v.is_watched).map(v => v.video_id)));
          const lastItem = videosRes.items[videosRes.items.length - 1];
          cursorRef.current = lastItem?.video_id;
        }

        if (playlistsRes) {
          setPlaylists(playlistsRes.items);
        }
      } finally {
        setIsLoading(false);
        setIsVideosLoading(false);
      }
    };
    load();
  }, [channelId, order, isAuthenticated, isAuthLoading]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasNextVideos) return;
    const uuid = channelUuidRef.current;
    if (!uuid) return;
    setIsLoadingMore(true);
    try {
      const res = await getChannel().getChannelsChannelIdVideos(uuid, { limit: PAGE_SIZES.CHANNEL_VIDEOS, cursor: cursorRef.current, order });
      setVideos(prev => [...prev, ...res.items]);
      setHasNextVideos(res.has_next);
      setWatchedVideoIds(prev => {
        const next = new Set(prev);
        for (const v of res.items) {
          if (v.is_watched) next.add(v.video_id);
        }
        return next;
      });
      const lastItem = res.items[res.items.length - 1];
      cursorRef.current = lastItem?.video_id;
    } finally {
      setIsLoadingMore(false);
    }
  }, [isLoadingMore, hasNextVideos, order]);

  const sentinelRef = useInfiniteScroll(loadMore);

  const watchedRef = useRef(watchedVideoIds);
  watchedRef.current = watchedVideoIds;
  const handleToggleWatched = useCallback(async (videoId: string) => {
    const isCurrentlyWatched = watchedRef.current.has(videoId);
    if (isCurrentlyWatched) {
      await getHistory().deleteVideosVideoIdWatched(videoId);
      setWatchedVideoIds(prev => {
        const next = new Set(prev);
        next.delete(videoId);
        return next;
      });
    } else {
      await getHistory().postVideosVideoIdWatched(videoId);
      setWatchedVideoIds(prev => new Set(prev).add(videoId));
    }
  }, []);

  const handleToggleSubscription = async () => {
    if (isToggling || !channelInfo) return;
    setIsToggling(true);
    try {
      if (isSubscribed) {
        await getChannel().deleteChannelsChannelIdSubscribe(channelInfo.channel_id);
        setIsSubscribed(false);
      } else {
        await getChannel().postChannelsSubscribe({
          channel_id: channelInfo.external_channel_custom_id,
        });
        setIsSubscribed(true);
      }
    } catch {
    } finally {
      setIsToggling(false);
    }
  };

  if (isLoading) {
    return (
      <DashboardLayout>
        <div class="flex-1 overflow-y-auto w-full max-w-[1200px] mx-auto px-6 py-6 lg:py-10">
          <ChannelInfoCardSkeleton />
          <div class="mb-8">
            <div class="flex gap-4 overflow-x-auto pb-2">
              <SkeletonRepeat count={4} render={(i) => <ChannelDetailPlaylistCardSkeleton key={i} />} />
            </div>
          </div>
          <div class="card-grid">
            <SkeletonRepeat count={6} render={(i) => <VideoCardSkeleton key={i} />} />
          </div>
        </div>
      </DashboardLayout>
    );
  }

  if (!channelInfo) {
    return (
      <DashboardLayout>
        <div class="w-full max-w-[1200px] mx-auto px-6 py-10">
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="search_off" class="text-5xl mb-4" />
            <p class="text-lg font-medium">{t("channelDetail.notFound")}</p>
            <BrowserBackLink
              class="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 no-underline"
              fallbackHref="/"
            >
              <Icon name="arrow_back" class="text-[18px]" />
              {t("channelDetail.backToDashboard")}
            </BrowserBackLink>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div class="flex-1 overflow-y-auto w-full max-w-[1200px] mx-auto px-6 py-6 lg:py-10">
        <ChannelInfoCard
          channelInfo={channelInfo}
          isSubscribed={isSubscribed}
          onToggleSubscription={() => requireAuth(handleToggleSubscription)}
          isToggling={isToggling}
        />

        {playlists.length > 0 && (
          <div class="mb-8">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-bold">
                {t("channelDetail.playlists")}
              </h3>
              <a
                href={`/channels/${channelId}/playlists`}
                class="text-sm font-medium text-primary hover:text-primary/80 no-underline"
              >
                {t("channelDetail.showMore")}
              </a>
            </div>
            <div class="flex gap-4 overflow-x-auto pb-2">
              {playlists.map((pl) => (
                <a
                  key={pl.playlist_id}
                  href={`/playlists/${pl.playlist_id}`}
                  class="group flex-shrink-0 w-56 bg-card-light dark:bg-card-dark rounded-xl border border-transparent hover:border-primary/20 overflow-hidden no-underline"
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
                    <h4 class="text-sm font-bold text-charcoal dark:text-white leading-tight line-clamp-2 group-hover:text-primary">
                      {pl.playlist_title}
                    </h4>
                    <span class="text-xs text-text-muted-light dark:text-text-muted-dark mt-1 block">
                      {t("playlists.videoCount", { count: pl.playlist_video_count })}
                    </span>
                  </div>
                </a>
              ))}
            </div>
          </div>
        )}

        <div>
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-bold">
              {t("channelDetail.latestUploads")}
            </h3>
            <select
              class="text-sm bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark rounded-lg px-3 py-1.5 cursor-pointer"
              value={order}
              onChange={(e) => {
                const nextOrder = parseChannelVideoOrder((e.target as HTMLSelectElement).value);
                if (!nextOrder) return;
                setOrder(nextOrder);
                setVideos([]);
                setHasNextVideos(false);
                cursorRef.current = undefined;
              }}
            >
              {CHANNEL_VIDEO_ORDER_OPTIONS.map((value) => (
                <option key={value} value={value}>
                  {t(`channelDetail.order${value === "newer" ? "Newer" : "Older"}`)}
                </option>
              ))}
            </select>
          </div>

          {isVideosLoading ? (
            <div class="card-grid">
              <SkeletonRepeat count={6} render={(i) => <VideoCardSkeleton key={i} />} />
            </div>
          ) : videos.length > 0 ? (
            <>
              <div class="card-grid">
                {videos.map((video) => (
                  <ChannelVideoCard
                    key={video.video_id}
                    video={video}
                    isWatched={watchedVideoIds.has(video.video_id)}
                    requireAuth={requireAuth}
                    onToggleWatched={handleToggleWatched}
                  />
                ))}
                {isLoadingMore && (
                  <SkeletonRepeat count={3} render={(i) => <VideoCardSkeleton key={`more-${i}`} />} />
                )}
              </div>
              <div ref={sentinelRef} class="h-1" />
              {!hasNextVideos && !isLoadingMore && videos.length > 0 && (
                <p class="text-center text-sm text-text-muted-light dark:text-text-muted-dark py-8">
                  {t("dashboard.endOfFeed")}
                </p>
              )}
            </>
          ) : (
            <div class="flex flex-col items-center justify-center py-12 text-text-muted-light dark:text-text-muted-dark bg-card-light dark:bg-card-dark rounded-xl border border-border-light dark:border-border-dark">
              <Icon name="videocam_off" class="text-4xl mb-3" />
              <p class="text-sm font-medium">{t("channelDetail.noVideos")}</p>
            </div>
          )}
        </div>
      </div>
      <AuthPromptDialog open={showAuthPrompt} onClose={closeAuthPrompt} />
    </DashboardLayout>
  );
}

export default function ChannelDetail({ channelId }: { channelId: string }) {
  return <ChannelDetailContent channelId={channelId} />;
}
