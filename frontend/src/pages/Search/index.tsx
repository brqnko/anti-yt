import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { VideoCard } from "../../components/VideoCard";
import { getFeed } from "../../api/generated/feed";
import { PAGE_SIZES } from "../../constants";
import type { GetSearch200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";

function SearchContent() {
  const { t } = useTranslation();
  const { query: urlQuery } = useLocation();
  const searchQuery = new URLSearchParams(urlQuery).get("q") || "";
  useTitle(searchQuery ? `${searchQuery} - ${t("search.pageTitle")}` : t("search.pageTitle"));

  const [videos, setVideos] = useState<GetSearch200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [error, setError] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);
  const currentQueryRef = useRef<string>("");

  useEffect(() => {
    if (!searchQuery.trim()) {
      setVideos([]);
      setIsLoading(false);
      setError(false);
      setHasNext(false);
      cursorRef.current = undefined;
      currentQueryRef.current = "";
      return;
    }

    currentQueryRef.current = searchQuery;
    setIsLoading(true);
    setError(false);
    setVideos([]);
    cursorRef.current = undefined;

    const load = async () => {
      try {
        const res = await getFeed().getSearch({ query: searchQuery, limit: PAGE_SIZES.SEARCH, language: navigator.language });
        if (currentQueryRef.current !== searchQuery) return;
        setVideos(res.items);
        setHasNext(res.has_next);
        cursorRef.current = res.cursor;
      } catch {
        if (currentQueryRef.current === searchQuery) setError(true);
      } finally {
        if (currentQueryRef.current === searchQuery) setIsLoading(false);
      }
    };
    load();
  }, [searchQuery]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasNext || !searchQuery.trim()) return;
    setIsLoadingMore(true);
    try {
      const res = await getFeed().getSearch({
        query: searchQuery,
        limit: PAGE_SIZES.SEARCH,
        cursor: cursorRef.current,
        language: navigator.language,
      });
      setVideos((prev) => [...prev, ...res.items]);
      setHasNext(res.has_next);
      cursorRef.current = res.cursor;
    } finally {
      setIsLoadingMore(false);
    }
  }, [isLoadingMore, hasNext, searchQuery]);

  const sentinelRef = useInfiniteScroll(loadMore);

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
        ) : isLoading ? (
          <LoadingSpinner />
        ) : error ? (
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="error_outline" class="text-5xl mb-4" />
            <p class="text-lg font-medium">{t("search.loadError")}</p>
            <button
              class="mt-4 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 transition-colors cursor-pointer border-none"
              onClick={() => window.location.reload()}
            >
              {t("search.retry")}
            </button>
          </div>
        ) : videos.length > 0 ? (
          <>
            <div class="flex flex-col gap-6">
              {videos.map((video) => (
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
                  layout="row"
                />
              ))}
            </div>
            <div ref={sentinelRef} class="h-1" />
            {isLoadingMore && <LoadingSpinner size="sm" className="py-8" />}
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
        )}
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
