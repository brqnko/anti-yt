import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { getChannel } from "../../api/generated/channel";
import { formatSubscriberCount } from "../../utils/format";
import type { GetChannelsSubscribed200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";

const PAGE_SIZE = 30;

function ChannelsContent() {
  const { t } = useTranslation();
  useTitle(t("channels.pageTitle"));

  const [channels, setChannels] = useState<GetChannelsSubscribed200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);

  const loadChannels = useCallback(async () => {
    setIsLoading(true);
    setError(false);
    try {
      const res = await getChannel().getChannelsSubscribed({ limit: PAGE_SIZE });
      setChannels(res.items);
      setHasNext(res.has_next);
      cursorRef.current = res.items[res.items.length - 1]?.channel_id;
    } catch {
      setError(true);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadChannels();
  }, [loadChannels]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasNext) return;
    setIsLoadingMore(true);
    try {
      const res = await getChannel().getChannelsSubscribed({
        limit: PAGE_SIZE,
        cursor: cursorRef.current,
      });
      setChannels((prev) => [...prev, ...res.items]);
      setHasNext(res.has_next);
      cursorRef.current = res.items[res.items.length - 1]?.channel_id;
    } finally {
      setIsLoadingMore(false);
    }
  }, [isLoadingMore, hasNext]);

  return (
    <DashboardLayout>
      <div class="flex-1 flex justify-center py-8 px-4 sm:px-8">
        <div class="w-full max-w-[1440px] flex flex-col gap-6">
          <h1 class="text-charcoal dark:text-white text-4xl font-black leading-tight tracking-[-0.033em]">
            {t("channels.pageTitle")}
          </h1>

          {/* Link to valuable channels */}
          <a
            href="/channels/explore"
            class="flex items-center gap-3 px-5 py-4 rounded-xl bg-primary/5 dark:bg-primary/10 border border-primary/20 hover:border-primary/40 transition-colors no-underline group"
          >
            <Icon name="recommend" class="text-2xl text-primary" />
            <div class="flex-1 min-w-0">
              <span class="text-base font-bold text-charcoal dark:text-white group-hover:text-primary transition-colors">
                {t("channels.exploreLink")}
              </span>
              <p class="text-sm text-text-muted-light dark:text-text-muted-dark mt-0.5">
                {t("channels.exploreLinkDesc")}
              </p>
            </div>
            <Icon name="chevron_right" class="text-xl text-text-muted-light dark:text-text-muted-dark group-hover:text-primary transition-colors" />
          </a>

          {/* Subscribed channels list */}
          {isLoading ? (
            <LoadingSpinner />
          ) : error ? (
            <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
              <Icon name="error_outline" class="text-5xl mb-4" />
              <p class="text-lg font-medium">{t("channels.loadError")}</p>
              <button
                class="mt-4 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 transition-colors cursor-pointer border-none"
                onClick={loadChannels}
              >
                {t("channels.retry")}
              </button>
            </div>
          ) : channels.length === 0 ? (
            <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
              <Icon name="subscriptions" class="text-5xl mb-4" />
              <p class="text-lg font-medium">{t("channels.noChannels")}</p>
            </div>
          ) : (
            <div class="flex flex-col gap-2">
              {channels.map((ch) => (
                <a
                  key={ch.channel_id}
                  href={`/channels/${ch.channel_id}`}
                  class="flex items-center gap-4 px-4 py-3 rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark hover:border-primary/50 dark:hover:border-primary/50 transition-colors no-underline"
                >
                  <img
                    alt={ch.external_channel_display_name}
                    loading="lazy"
                    class="size-11 rounded-full object-cover bg-gray-100 shrink-0"
                    src={ch.external_channel_icon_url}
                  />
                  <div class="flex-1 min-w-0">
                    <p class="font-bold text-charcoal dark:text-white text-sm leading-snug truncate">
                      {ch.external_channel_display_name}
                    </p>
                    <p class="text-xs text-text-muted-light dark:text-text-muted-dark mt-0.5">
                      {ch.channel_custom_id}
                      {" · "}
                      {formatSubscriberCount(ch.channel_subscribers_count)}{" "}
                      {t("channelDetail.subscribers")}
                    </p>
                  </div>
                  <Icon name="chevron_right" class="text-xl text-text-muted-light dark:text-text-muted-dark shrink-0" />
                </a>
              ))}
              {hasNext && (
                <button
                  class="mt-4 self-center px-6 py-2.5 bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark rounded-lg text-sm font-medium text-charcoal dark:text-white hover:border-primary/30 transition-colors cursor-pointer"
                  onClick={loadMore}
                  disabled={isLoadingMore}
                >
                  {isLoadingMore ? <LoadingSpinner /> : t("channels.loadMore")}
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
}

export default function Channels() {
  return (
    <ProtectedRoute>
      <ChannelsContent />
    </ProtectedRoute>
  );
}
