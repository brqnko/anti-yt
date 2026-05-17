import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { useMeta } from "../../hooks/useMeta";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { VideoCard } from "../../components/VideoCard";
import { getFeed } from "../../api/generated/feed";
import { PAGE_SIZES } from "../../constants";
import type { GetSearch200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";
import { VideoCardSkeleton, ChannelRowSkeleton, SkeletonRepeat } from "../../components/skeletons";
import { formatSubscriberCount } from "../../utils/format";

function ChannelRow({ item }: { item: GetSearch200ItemsItem }) {
  const { t } = useTranslation();
  return (
    <a
      href={`/channels/${item.channel_id}`}
      class="flex items-center gap-4 p-3 rounded-lg bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark hover:border-primary/50 no-underline text-inherit"
    >
      <img
        src={item.external_channel_icon_url}
        alt=""
        loading="lazy"
        class="rounded-full size-12 shrink-0 border border-border-light dark:border-border-dark object-cover"
      />
      <div class="flex flex-col grow min-w-0">
        <p class="font-bold truncate text-charcoal dark:text-white">
          {item.external_channel_display_name}
        </p>
        <p class="text-xs text-text-muted-light dark:text-text-muted-dark">
          {item.channel_custom_id}
          {item.channel_subscribers_count != null && (
            <>
              {" · "}
              {formatSubscriberCount(item.channel_subscribers_count)}{" "}
              {t("channelDetail.subscribers")}
            </>
          )}
        </p>
      </div>
    </a>
  );
}

function SearchContent() {
  const { t } = useTranslation();
  const { query: urlQuery } = useLocation();
  const params = new URLSearchParams(urlQuery);
  const searchQuery = params.get("q") || "";
  useMeta({
    title: searchQuery ? `${searchQuery} - ${t("search.pageTitle")}` : t("search.pageTitle"),
    description: t("search.metaDescription"),
    canonicalPath: "/search",
  });

  const order = params.get("order") || undefined;
  const published_after = params.get("published_after") || undefined;
  const published_before = params.get("published_before") || undefined;
  const region_code = params.get("region_code") || undefined;
  const relevance_language = params.get("relevance_language") || undefined;
  const typeParam = params.get("type") as "channel" | "video" | null;

  const filterKey = JSON.stringify({ order, published_after, published_before, region_code, relevance_language });

  const [items, setItems] = useState<GetSearch200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [error, setError] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);
  const currentQueryRef = useRef<string>("");

  const buildParams = useCallback(
    (cursor?: string) => ({
      query: searchQuery,
      limit: PAGE_SIZES.SEARCH,
      language: navigator.language,
      cursor,
      order: order as any,
      published_after: published_after ? `${published_after}T00:00:00Z` : undefined,
      published_before: published_before ? `${published_before}T23:59:59Z` : undefined,
      region_code,
      relevance_language,
    }),
    [searchQuery, order, published_after, published_before, region_code, relevance_language],
  );

  useEffect(() => {
    if (!searchQuery.trim()) {
      setItems([]);
      setIsLoading(false);
      setError(false);
      setHasNext(false);
      cursorRef.current = undefined;
      currentQueryRef.current = "";
      return;
    }

    const isQueryChange = currentQueryRef.current !== searchQuery;
    currentQueryRef.current = searchQuery;
    setIsLoading(true);
    setError(false);
    if (isQueryChange) {
      setItems([]);
    }
    cursorRef.current = undefined;

    const load = async () => {
      try {
        const res = await getFeed().getSearch(buildParams());
        if (currentQueryRef.current !== searchQuery) return;
        setItems(res.items);
        setHasNext(res.has_next);
        cursorRef.current = res.cursor;
      } catch {
        if (currentQueryRef.current === searchQuery) setError(true);
      } finally {
        if (currentQueryRef.current === searchQuery) setIsLoading(false);
      }
    };
    load();
  }, [searchQuery, filterKey]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasNext || !searchQuery.trim()) return;
    setIsLoadingMore(true);
    try {
      const res = await getFeed().getSearch(buildParams(cursorRef.current));
      setItems((prev) => [...prev, ...res.items]);
      setHasNext(res.has_next);
      cursorRef.current = res.cursor;
    } finally {
      setIsLoadingMore(false);
    }
  }, [isLoadingMore, hasNext, searchQuery, buildParams]);

  const sentinelRef = useInfiniteScroll(loadMore);

  const renderItem = (item: GetSearch200ItemsItem) => {
    if (item.type === "channel") {
      return <ChannelRow key={item.channel_id} item={item} />;
    }
    return (
      <VideoCard
        key={item.video_id}
        videoId={item.video_id!}
        thumbnailUrl={item.external_video_thumbnail_url!}
        title={item.external_video_title!}
        lengthSeconds={item.external_video_length_seconds!}
        channel={{
          channelId: item.channel_id,
          iconUrl: item.external_channel_icon_url,
          displayName: item.external_channel_display_name,
        }}
        dateStr={item.external_video_created_at!}
        watchedSeconds={item.last_watch_seconds}
        layout="row"
      />
    );
  };

  return (
    <DashboardLayout>
      <div class="p-6">
        {searchQuery.trim() && (
          <h1 class="text-xl font-bold text-charcoal dark:text-white mb-6">
            {t("search.resultsFor", { query: searchQuery })}
          </h1>
        )}

        {!searchQuery.trim() ? (
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="search" class="text-5xl mb-4" />
            <p class="text-lg font-medium">{t("search.placeholder")}</p>
          </div>
        ) : isLoading && items.length === 0 ? (
          <div class="flex flex-col gap-4">
            <SkeletonRepeat count={2} render={(i) => <ChannelRowSkeleton key={`ch-${i}`} />} />
            <SkeletonRepeat count={4} render={(i) => <VideoCardSkeleton key={i} layout="row" />} />
          </div>
        ) : error ? (
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="error_outline" class="text-5xl mb-4" />
            <p class="text-lg font-medium">{t("search.loadError")}</p>
            <button
              class="mt-4 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 cursor-pointer border-none"
              onClick={() => window.location.reload()}
            >
              {t("search.retry")}
            </button>
          </div>
        ) : (() => {
          const filteredItems = typeParam ? items.filter((item) => item.type === typeParam) : items;
          return filteredItems.length > 0 ? (
            <>
              <div class="flex flex-col gap-4">
                {filteredItems.map(renderItem)}
                {isLoadingMore && (
                  <SkeletonRepeat count={3} render={(i) => <VideoCardSkeleton key={`more-${i}`} layout="row" />} />
                )}
              </div>
              <div ref={sentinelRef} class="h-1" />
              {!hasNext && !isLoadingMore && (
                <p class="text-center text-sm text-text-muted-light dark:text-text-muted-dark py-8">
                  {t("search.endOfResults")}
                </p>
              )}
            </>
          ) : (
            <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
              <Icon name="search_off" class="text-5xl mb-4" />
              <p class="text-lg font-medium">{t("search.noResults")}</p>
              <p class="text-sm mt-1">{t("search.noResultsDesc")}</p>
            </div>
          );
        })()}
      </div>
    </DashboardLayout>
  );
}

export default function Search() {
  return (
    <ProtectedRoute>
      <SearchContent />
    </ProtectedRoute>
  );
}
