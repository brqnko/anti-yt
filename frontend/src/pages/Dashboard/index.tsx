import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { getVideo } from "../../api/generated/video";
import { formatDuration, formatTimeAgo } from "../../utils/format";
import type { GetFeed200ItemsItem } from "../../api/generated/antiYtApi.schemas";

function VideoCard({
  video,
  t,
}: {
  video: GetFeed200ItemsItem;
  t: (key: string, opts?: object) => string;
}) {
  return (
    <article class="flex flex-col gap-3">
      <a href={`/watch/${video.video_id}`} class="group/thumb relative aspect-video rounded-xl overflow-hidden bg-gray-200 dark:bg-gray-800 cursor-pointer block no-underline">
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
      </a>
      <div class="flex gap-3 items-start">
        <a href={`/channels/${video.channel_id}`} class="size-9 rounded-full bg-gray-300 dark:bg-gray-700 flex-shrink-0 overflow-hidden cursor-pointer">
          <img
            alt={video.external_channel_display_name}
            class="w-full h-full object-cover"
            src={video.external_channel_icon_url}
          />
        </a>
        <div class="flex flex-col min-w-0 flex-1">
          <a href={`/watch/${video.video_id}`} class="text-base font-bold text-charcoal dark:text-white leading-tight line-clamp-2 cursor-pointer no-underline hover:text-primary transition-colors">
            {video.external_video_title}
          </a>
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

function DashboardContent() {
  const { t } = useTranslation();
  useTitle(t("dashboard.pageTitle"));

  const [feedVideos, setFeedVideos] = useState<GetFeed200ItemsItem[]>([]);
  const [isLoadingFeed, setIsLoadingFeed] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    const loadData = async () => {
      try {
        const feedRes = await getVideo().getFeed({ limit: 12 });
        setFeedVideos(feedRes.items);
        setHasNext(feedRes.has_next);
        const lastItem = feedRes.items[feedRes.items.length - 1];
        cursorRef.current = lastItem?.video_id;
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
    <DashboardLayout>
      <div class="p-6">
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
