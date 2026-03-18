import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { getChannel } from "../../api/generated/channel";
import { formatSubscriberCount } from "../../utils/format";
import { VideoCard } from "../../components/VideoCard";
import type {
  GetChannelsChannelIdVideos200ItemsItem,
} from "../../api/generated/antiYtApi.schemas";
import { Linkify } from "../../components/Linkify";

interface ChannelInfo {
  channel_id: string;
  external_channel_id: string;
  display_name: string;
  description?: string;
  icon_url: string;
  subscribers_count?: number;
  custom_id?: string;
}

function ExpandableDescription({ description }: { description: string }) {
  const { t } = useTranslation();
  const [expanded, setExpanded] = useState(false);
  const [clamped, setClamped] = useState(false);
  const ref = useRef<HTMLParagraphElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (el) setClamped(el.scrollHeight > el.clientHeight);
  }, [description]);

  return (
    <>
      <div class="h-px bg-border-light dark:bg-border-dark my-5" />
      <p
        ref={ref}
        class={`text-sm text-text-muted-light dark:text-text-muted-dark leading-relaxed whitespace-pre-line ${expanded ? "" : "line-clamp-3"}`}
      >
        <Linkify text={description} />
      </p>
      {clamped && (
        <button
          type="button"
          class="text-sm font-medium text-primary hover:text-primary/80 transition-colors bg-transparent border-none cursor-pointer p-0 mt-2"
          onClick={() => { setExpanded(!expanded); if (expanded && ref.current) setClamped(ref.current.scrollHeight > ref.current.clientHeight); }}
        >
          {expanded ? t("channelDetail.showLess") : t("channelDetail.showMore")}
        </button>
      )}
    </>
  );
}

function ChannelDetailContent({ channelId }: { channelId: string }) {
  const { t } = useTranslation();

  const [channelInfo, setChannelInfo] = useState<ChannelInfo | null>(null);
  const [videos, setVideos] = useState<GetChannelsChannelIdVideos200ItemsItem[]>([]);
  const [isSubscribed, setIsSubscribed] = useState(false);
  const [subscriptionId, setSubscriptionId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isToggling, setIsToggling] = useState(false);
  const [hasNextVideos, setHasNextVideos] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);

  useTitle(channelInfo?.display_name ?? t("channelDetail.pageTitle"));

  useEffect(() => {
    const load = async () => {
      try {
        const [subsRes, videosRes] = await Promise.allSettled([
          getChannel().getSubscriptions({ limit: 50 }),
          getChannel().getChannelsChannelIdVideos(channelId),
        ]);

        if (subsRes.status === "fulfilled") {
          const found = subsRes.value.items.find(s => s.channel_id === channelId);
          if (found) {
            setChannelInfo({
              channel_id: found.channel_id,
              external_channel_id: found.external_channel_id,
              display_name: found.external_channel_display_name,
              description: found.channel_description,
              icon_url: found.external_channel_icon_url,
              subscribers_count: found.channel_subscribers_count,
              custom_id: found.channel_custom_id,
            });
            setIsSubscribed(true);
            setSubscriptionId(found.subscription_id);
          }
        }

        if (videosRes.status === "fulfilled") {
          setVideos(videosRes.value.items);
          setHasNextVideos(videosRes.value.has_next);
          const lastItem = videosRes.value.items[videosRes.value.items.length - 1];
          cursorRef.current = lastItem?.video_id;
        }
      } finally {
        setIsLoading(false);
      }
    };
    load();
  }, [channelId]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasNextVideos) return;
    setIsLoadingMore(true);
    try {
      const res = await getChannel().getChannelsChannelIdVideos(channelId, { cursor: cursorRef.current });
      setVideos(prev => [...prev, ...res.items]);
      setHasNextVideos(res.has_next);
      const lastItem = res.items[res.items.length - 1];
      cursorRef.current = lastItem?.video_id;
    } finally {
      setIsLoadingMore(false);
    }
  }, [channelId, isLoadingMore, hasNextVideos]);

  const sentinelRef = useInfiniteScroll(loadMore);

  const handleToggleSubscription = async () => {
    if (isToggling || !channelInfo) return;
    setIsToggling(true);
    try {
      if (isSubscribed && subscriptionId) {
        await getChannel().deleteSubscriptionsSubscriptionId(subscriptionId);
        setIsSubscribed(false);
        setSubscriptionId(null);
      } else {
        const result = await getChannel().postSubscriptions({
          channel_id: channelInfo.external_channel_id,
        });
        setChannelInfo({
          channel_id: result.channel_id,
          external_channel_id: result.external_channel_id,
          display_name: result.external_channel_display_name,
          description: result.channel_description,
          icon_url: result.external_channel_icon_url,
          subscribers_count: result.channel_subscribers_count,
          custom_id: result.channel_custom_id,
        });
        setSubscriptionId(result.subscription_id);
        setIsSubscribed(true);
      }
    } catch {
      // silently fail
    } finally {
      setIsToggling(false);
    }
  };

  if (isLoading) {
    return (
      <DashboardLayout>
        <LoadingSpinner className="py-32" />
      </DashboardLayout>
    );
  }

  if (!channelInfo) {
    return (
      <DashboardLayout>
        <div class="w-full max-w-[1200px] mx-auto px-6 py-10">
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <span class="material-symbols-outlined text-5xl mb-4">search_off</span>
            <p class="text-lg font-medium">{t("channelDetail.notFound")}</p>
            <p class="text-sm mt-1">{t("channelDetail.notFoundDesc")}</p>
            <a
              class="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 transition-colors no-underline"
              href="/dashboard"
            >
              <span class="material-symbols-outlined text-[18px]">arrow_back</span>
              {t("channelDetail.backToDashboard")}
            </a>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div class="flex-1 overflow-y-auto w-full max-w-[1200px] mx-auto px-6 py-6 lg:py-10">
        {/* Channel Info */}
        <div class="bg-card-light dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark mb-8 p-6">
          <div class="flex flex-col md:flex-row gap-6 items-start md:items-center">
            {/* Avatar */}
            <div class="shrink-0">
              <div class="size-24 md:size-28 rounded-full bg-gray-200 dark:bg-gray-800 overflow-hidden shadow-md border-2 border-border-light dark:border-border-dark">
                <img
                  src={channelInfo.icon_url}
                  alt={channelInfo.display_name}
                  class="w-full h-full object-cover"
                />
              </div>
            </div>

            {/* Channel Text Info */}
            <div class="flex-1 min-w-0">
              <h1 class="text-2xl md:text-3xl font-bold mb-1">{channelInfo.display_name}</h1>
              <div class="flex flex-wrap gap-x-4 gap-y-1 text-sm text-text-muted-light dark:text-text-muted-dark">
                {channelInfo.subscribers_count != null && (
                  <span>
                    {formatSubscriberCount(channelInfo.subscribers_count)} {t("channelDetail.subscribers")}
                  </span>
                )}
                {channelInfo.custom_id && (
                  <span>{channelInfo.custom_id}</span>
                )}
              </div>
            </div>

            {/* Whitelist Toggle */}
            <div class="flex-shrink-0 w-full md:w-auto">
              <div class="bg-background-light dark:bg-background-dark border border-primary/20 p-4 rounded-xl flex items-center justify-between md:justify-start gap-4 shadow-sm">
                <div class="flex flex-col">
                  <span class="text-sm font-bold">{t("channelDetail.whitelistChannel")}</span>
                  <span class="text-xs text-text-muted-light dark:text-text-muted-dark">
                    {t("channelDetail.whitelistDesc")}
                  </span>
                </div>
                <button
                  class="relative inline-flex items-center cursor-pointer bg-transparent border-none p-0"
                  onClick={handleToggleSubscription}
                  disabled={isToggling}
                >
                  <div
                    class={`w-14 h-7 rounded-full transition-colors duration-200 ${
                      isSubscribed ? "bg-primary" : "bg-gray-200 dark:bg-gray-700"
                    } ${isToggling ? "opacity-50" : ""}`}
                  >
                    <div
                      class={`absolute top-0.5 left-[4px] bg-white border border-gray-300 rounded-full h-6 w-6 transition-transform duration-200 ${
                        isSubscribed ? "translate-x-full" : ""
                      }`}
                    />
                  </div>
                </button>
              </div>
            </div>
          </div>

          {/* Description */}
          {channelInfo.description != null && (
            <ExpandableDescription description={channelInfo.description} />
          )}
        </div>

        {/* Latest Uploads */}
        <div>
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-bold">
              {t("channelDetail.latestUploads")}
              <span class="text-text-muted-light dark:text-text-muted-dark font-normal text-sm ml-2">
                ({t("channelDetail.previewOnly")})
              </span>
            </h3>
          </div>

          {videos.length > 0 ? (
            <>
              <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
                {videos.map((video) => (
                  <VideoCard
                    key={video.video_id}
                    videoId={video.video_id}
                    thumbnailUrl={video.external_video_thumbnail_url}
                    title={video.external_video_title}
                    lengthSeconds={video.external_video_length_seconds}
                    dateStr={video.external_video_created_at}
                    watchedSeconds={video.last_watch_seconds}
                  />
                ))}
              </div>
              <div ref={sentinelRef} class="h-1" />
              {isLoadingMore && <LoadingSpinner size="sm" className="py-8" />}
              {!hasNextVideos && !isLoadingMore && videos.length > 0 && (
                <p class="text-center text-sm text-text-muted-light dark:text-text-muted-dark py-8">
                  🎉 {t("dashboard.endOfFeed")}
                </p>
              )}
            </>
          ) : (
            <div class="flex flex-col items-center justify-center py-12 text-text-muted-light dark:text-text-muted-dark bg-card-light dark:bg-card-dark rounded-xl border border-border-light dark:border-border-dark">
              <span class="material-symbols-outlined text-4xl mb-3">videocam_off</span>
              <p class="text-sm font-medium">{t("channelDetail.noVideos")}</p>
            </div>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
}

export default function ChannelDetail({ channelId }: { channelId: string }) {
  return (
    <ProtectedRoute>
      <ChannelDetailContent channelId={channelId} />
    </ProtectedRoute>
  );
}
