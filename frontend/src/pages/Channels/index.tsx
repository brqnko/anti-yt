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
import { Dialog } from "../../components/Dialog";
import { ExploreChannelsBanner } from "../../components/ExploreChannelsBanner";

const PAGE_SIZE = 30;

function RemoveChannelDialog({
  open,
  channel,
  onClose,
  onConfirm,
}: {
  open: boolean;
  channel: GetChannelsSubscribed200ItemsItem | null;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}) {
  const { t } = useTranslation();
  const [isRemoving, setIsRemoving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) {
      setIsRemoving(false);
      setError(null);
    }
  }, [open]);

  if (!open || !channel) return null;

  const handleConfirm = async () => {
    if (isRemoving) return;
    setIsRemoving(true);
    setError(null);
    try {
      await onConfirm();
      onClose();
    } catch {
      setError(t("channels.unsubscribeError"));
    } finally {
      setIsRemoving(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} ariaLabel={t("channels.unsubscribeDialog.title")} maxWidth="max-w-sm" showCloseButton>
        <div class="flex items-center gap-3 mb-4">
          <h2 class="text-lg font-bold text-charcoal dark:text-white">
            {t("channels.unsubscribeDialog.title")}
          </h2>
        </div>
        <div class="flex items-center gap-3 p-3 rounded-lg bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark mb-4">
          <img
            src={channel.external_channel_icon_url}
            alt=""
            loading="lazy"
            class="rounded-full size-10 shrink-0 border border-border-light dark:border-border-dark object-cover"
          />
          <div class="flex flex-col min-w-0">
            <p class="font-bold truncate text-sm">{channel.external_channel_display_name}</p>
            <p class="text-xs text-text-muted-light dark:text-text-muted-dark">{channel.channel_custom_id}</p>
          </div>
        </div>
        <p class="text-sm text-text-muted-light dark:text-text-muted-dark mb-4">
          {t("channels.unsubscribeDialog.description", { name: channel.external_channel_display_name })}
        </p>
        {error && (
          <p class="text-sm text-red-500 mb-4">{error}</p>
        )}
        <div class="flex justify-end gap-3">
          <button
            class="px-4 py-2 rounded-xl text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 bg-transparent border-none cursor-pointer"
            onClick={onClose}
          >
            {t("channels.unsubscribeDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-xl text-sm font-bold text-white bg-red-500 hover:bg-red-600 border-none cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={isRemoving}
            onClick={handleConfirm}
          >
            {isRemoving
              ? t("channels.unsubscribeDialog.removing")
              : t("channels.unsubscribeDialog.remove")}
          </button>
        </div>
    </Dialog>
  );
}

function ChannelsContent() {
  const { t } = useTranslation();
  useTitle(t("channels.pageTitle"));

  const [channels, setChannels] = useState<GetChannelsSubscribed200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [removeTarget, setRemoveTarget] = useState<GetChannelsSubscribed200ItemsItem | null>(null);
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

  const handleRemoveConfirm = useCallback(async () => {
    if (!removeTarget) return;
    await getChannel().deleteChannelsChannelIdSubscribe(removeTarget.channel_id);
    setChannels((prev) => prev.filter((ch) => ch.channel_id !== removeTarget.channel_id));
  }, [removeTarget]);

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
          <ExploreChannelsBanner />

          {/* Subscribed channels list */}
          {isLoading ? (
            <LoadingSpinner />
          ) : error ? (
            <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
              <Icon name="error_outline" class="text-5xl mb-4" />
              <p class="text-lg font-medium">{t("channels.loadError")}</p>
              <button
                class="mt-4 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 cursor-pointer border-none"
                onClick={loadChannels}
              >
                {t("channels.retry")}
              </button>
            </div>
          ) : channels.length === 0 ? (
            <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
              <p class="text-lg font-medium">{t("channels.noChannels")}</p>
            </div>
          ) : (
            <div class="flex flex-col gap-3">
              {channels.map((ch) => (
                <div
                  key={ch.channel_id}
                  class="flex items-center gap-4 p-3 rounded-lg bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark group hover:border-primary/50"
                >
                  <a
                    href={`/channels/${ch.channel_id}`}
                    class="flex items-center gap-4 grow min-w-0 no-underline text-inherit"
                  >
                    <img
                      src={ch.external_channel_icon_url}
                      alt=""
                      loading="lazy"
                      class="rounded-full size-12 shrink-0 border border-border-light dark:border-border-dark object-cover"
                    />
                    <div class="flex flex-col grow min-w-0">
                      <p class="font-bold truncate">
                        {ch.external_channel_display_name}
                      </p>
                      <p class="text-xs text-text-muted-light dark:text-text-muted-dark">
                        {ch.channel_custom_id}
                        {" · "}
                        {formatSubscriberCount(ch.channel_subscribers_count)}{" "}
                        {t("channelDetail.subscribers")}
                      </p>
                    </div>
                  </a>
                  <button
                    class="size-8 flex items-center justify-center rounded-full text-text-muted-light dark:text-text-muted-dark hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 cursor-pointer bg-transparent border-none shrink-0"
                    onClick={() => setRemoveTarget(ch)}
                  >
                    <Icon name="close" class="text-[20px]" />
                  </button>
                </div>
              ))}
              {hasNext && (
                <button
                  class="mt-4 self-center px-6 py-2.5 bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark rounded-lg text-sm font-medium text-charcoal dark:text-white hover:border-primary/30 cursor-pointer"
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

      <RemoveChannelDialog
        open={!!removeTarget}
        channel={removeTarget}
        onClose={() => setRemoveTarget(null)}
        onConfirm={handleRemoveConfirm}
      />
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
