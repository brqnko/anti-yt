import { useState, useEffect, useCallback, useRef, useMemo } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { VideoCard } from "../../components/VideoCard";
import { getHistory } from "../../api/generated/history";
import { isoToDateStr, toDateStr, today } from "../../utils/format";
import { PAGE_SIZES } from "../../constants";
import type { GetHistory200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";
import { VideoCardSkeleton, SkeletonRepeat } from "../../components/skeletons";

function formatDateHeader(
  dateKey: string,
  t: (key: string) => string,
  locale: string,
): string {
  const td = today();
  const todayKey = toDateStr(td);

  const yesterday = new Date(td.getTime() - 86400000);
  const yesterdayKey = toDateStr(yesterday);

  if (dateKey === todayKey) return t("history.today");
  if (dateKey === yesterdayKey) return t("history.yesterday");

  const [y, m, d] = dateKey.split("-").map(Number);
  const date = new Date(y, m - 1, d);
  return date.toLocaleDateString(locale, {
    month: "long",
    day: "numeric",
    ...(y !== td.getFullYear() ? { year: "numeric" } : {}),
  });
}

export function HistoryTab() {
  const { t, i18n } = useTranslation();

  const [items, setItems] = useState<GetHistory200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [error, setError] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);
  const hasNextRef = useRef(false);
  const loadingMoreRef = useRef(false);

  const loadInitial = useCallback(async () => {
    setIsLoading(true);
    setError(false);
    try {
      const res = await getHistory().getHistory({ limit: PAGE_SIZES.HISTORY });
      setItems(res.items);
      setHasNext(res.has_next);
      hasNextRef.current = res.has_next;
      cursorRef.current = res.items[res.items.length - 1]?.watch_id;
    } catch {
      setError(true);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadInitial();
  }, [loadInitial]);

  const loadMore = useCallback(async () => {
    if (loadingMoreRef.current || !hasNextRef.current) return;
    loadingMoreRef.current = true;
    setIsLoadingMore(true);
    try {
      const res = await getHistory().getHistory({
        limit: 20,
        cursor: cursorRef.current,
      });
      setItems((prev) => [...prev, ...res.items]);
      setHasNext(res.has_next);
      hasNextRef.current = res.has_next;
      cursorRef.current = res.items[res.items.length - 1]?.watch_id;
    } catch {
    } finally {
      loadingMoreRef.current = false;
      setIsLoadingMore(false);
    }
  }, []);

  const groupedItems = useMemo(() => {
    const groups: { key: string; label: string; items: GetHistory200ItemsItem[] }[] = [];
    let currentKey = "";
    for (const item of items) {
      const key = isoToDateStr(item.watched_at);
      if (key !== currentKey) {
        currentKey = key;
        groups.push({
          key,
          label: formatDateHeader(key, t, i18n.language),
          items: [item],
        });
      } else {
        groups[groups.length - 1].items.push(item);
      }
    }
    return groups;
  }, [items, t, i18n.language]);

  const sentinelRef = useInfiniteScroll(loadMore);

  if (isLoading) {
    return (
      <div class="flex flex-col divide-y divide-gray-200 dark:divide-gray-800">
        <SkeletonRepeat
          count={5}
          render={(i) => (
            <div key={i} class="py-4 first:pt-0">
              <VideoCardSkeleton layout="row" />
            </div>
          )}
        />
      </div>
    );
  }

  return (
    <div class="flex flex-col gap-8">
      {error ? (
        <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
          <Icon name="error_outline" class="text-5xl mb-4" />
          <p class="text-lg font-medium">{t("history.loadError")}</p>
          <button
            onClick={loadInitial}
            class="mt-4 text-sm text-primary hover:underline cursor-pointer bg-transparent border-none"
          >
            {t("history.retry")}
          </button>
        </div>
      ) : items.length > 0 ? (
        <>
          <div class="flex flex-col gap-8">
            {groupedItems.map((group) => (
              <section key={group.key}>
                <h2 class="text-2xl font-bold text-charcoal dark:text-white mb-3">
                  {group.label}
                </h2>
                <div class="flex flex-col divide-y divide-gray-200 dark:divide-gray-800">
                  {group.items.map((item) => (
                    <div key={item.watch_id} class="py-4 first:pt-0">
                      <VideoCard
                        layout="row"
                        videoId={item.video_id}
                        thumbnailUrl={item.external_video_thumbnail_url}
                        title={item.external_video_title}
                        lengthSeconds={item.external_video_length_seconds}
                        channel={{
                          channelId: item.channel_id,
                          iconUrl: item.external_channel_icon_url,
                          displayName: item.external_channel_display_name,
                        }}
                        watchedAt={item.watched_at}
                        watchedSeconds={item.watch_position_seconds}
                        isWatched
                      />
                    </div>
                  ))}
                </div>
              </section>
            ))}
          </div>
          {isLoadingMore && (
            <div class="flex flex-col divide-y divide-gray-200 dark:divide-gray-800">
              <SkeletonRepeat
                count={3}
                render={(i) => (
                  <div key={`more-${i}`} class="py-4 first:pt-0">
                    <VideoCardSkeleton layout="row" />
                  </div>
                )}
              />
            </div>
          )}
          {hasNext && <div ref={sentinelRef} class="h-1" />}
          {!hasNext && !isLoadingMore && (
            <p class="text-center text-sm text-text-muted-light dark:text-text-muted-dark py-8">
              🎉 {t("history.endOfHistory")}
            </p>
          )}
        </>
      ) : (
        <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
          <Icon name="history" class="text-5xl mb-4" />
          <p class="text-lg font-medium">{t("history.noHistory")}</p>
        </div>
      )}
    </div>
  );
}
