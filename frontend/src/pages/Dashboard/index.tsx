import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { VideoCard } from "../../components/VideoCard";
import { getFeed } from "../../api/generated/feed";
import { getChannel } from "../../api/generated/channel";
import { PAGE_SIZES } from "../../constants";
import type { GetFeed200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";

function DashboardContent() {
  const { t } = useTranslation();
  useTitle(t("dashboard.pageTitle"));

  const [feedVideos, setFeedVideos] = useState<GetFeed200ItemsItem[]>([]);
  const [isLoadingFeed, setIsLoadingFeed] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [subscribedChannelIds, setSubscribedChannelIds] = useState<Set<string>>(new Set());
  const cursorRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [feedRes, subsRes] = await Promise.allSettled([
          getFeed().getFeed({ limit: PAGE_SIZES.FEED }),
          getChannel().getChannelsSubscribed({ limit: 50 }),
        ]);
        if (feedRes.status === "fulfilled") {
          setFeedVideos(feedRes.value.items);
          setHasNext(feedRes.value.has_next);
          const lastItem = feedRes.value.items[feedRes.value.items.length - 1];
          cursorRef.current = lastItem?.video_id;
        }
        if (subsRes.status === "fulfilled") {
          setSubscribedChannelIds(new Set(subsRes.value.items.map((s) => s.channel_id)));
        }
      } finally {
        setIsLoadingFeed(false);
      }
    };
    loadData();
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

  return (
    <DashboardLayout>
      <div class="p-6">
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
                  onToggleSubscription={() => handleToggleSubscription(video.channel_id)}
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
              href="/channels"
            >
              <Icon name="recommend" class="text-[18px]" />
              {t("dashboard.nav.recommendedChannels")}
            </a>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}

export default function Dashboard() {
  return (
    <ProtectedRoute>
      <DashboardContent />
    </ProtectedRoute>
  );
}
