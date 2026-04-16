import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { useRequireAuth } from "../../hooks/useRequireAuth";
import { DashboardLayout } from "../../components/DashboardLayout";
import { AuthPromptDialog } from "../../components/AuthPromptDialog";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { ChannelInfoCard } from "../../components/ChannelInfoCard";
import { getChannel } from "../../api/generated/channel";
import { PAGE_SIZES } from "../../constants";
import type {
  GetChannelsChannelId200,
  GetChannelsChannelIdPlaylists200ItemsItem,
} from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";

function ChannelPlaylistsContent({ channelId }: { channelId: string }) {
  const { t } = useTranslation();
  const { isAuthenticated, isLoading: isAuthLoading, requireAuth, showAuthPrompt, closeAuthPrompt } = useRequireAuth();

  const [channelInfo, setChannelInfo] = useState<GetChannelsChannelId200 | null>(null);
  const [playlists, setPlaylists] = useState<GetChannelsChannelIdPlaylists200ItemsItem[]>([]);
  const [isSubscribed, setIsSubscribed] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [isToggling, setIsToggling] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);

  useTitle(
    channelInfo
      ? `${channelInfo.external_channel_display_name} - ${t("channelDetail.playlists")}`
      : t("channelDetail.playlists")
  );

  useEffect(() => {
    if (isAuthLoading) return;
    const load = async () => {
      setIsLoading(true);
      try {
        const [channelRes, playlistsRes, subsRes] = await Promise.all([
          getChannel().getChannelsChannelId(channelId).catch(() => null),
          getChannel().getChannelsChannelIdPlaylists(channelId, { limit: PAGE_SIZES.CHANNEL_PLAYLISTS_PAGE }).catch(() => null),
          isAuthenticated
            ? getChannel().getChannelsSubscribed({ limit: 50 }).catch(() => null)
            : Promise.resolve(null),
        ]);

        if (channelRes) {
          setChannelInfo(channelRes);
        }

        if (playlistsRes) {
          setPlaylists(playlistsRes.items);
          setHasNext(playlistsRes.has_next);
          const last = playlistsRes.items[playlistsRes.items.length - 1];
          cursorRef.current = last?.playlist_id;
        }

        if (subsRes) {
          const found = subsRes.items.find(s => s.channel_id === channelId);
          if (found) {
            setIsSubscribed(true);
          }
        }
      } finally {
        setIsLoading(false);
      }
    };
    load();
  }, [channelId, isAuthenticated, isAuthLoading]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasNext) return;
    setIsLoadingMore(true);
    try {
      const res = await getChannel().getChannelsChannelIdPlaylists(channelId, {
        limit: PAGE_SIZES.CHANNEL_PLAYLISTS_PAGE,
        cursor: cursorRef.current,
      });
      setPlaylists((prev) => [...prev, ...res.items]);
      setHasNext(res.has_next);
      const last = res.items[res.items.length - 1];
      cursorRef.current = last?.playlist_id;
    } finally {
      setIsLoadingMore(false);
    }
  }, [channelId, isLoadingMore, hasNext]);

  const sentinelRef = useInfiniteScroll(loadMore);

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

  return (
    <DashboardLayout>
      <div class="flex-1 overflow-y-auto w-full max-w-[1200px] mx-auto px-6 py-6 lg:py-10">
        <div class="flex items-center gap-3 mb-6">
          <a
            href={`/channels/${channelId}`}
            class="inline-flex items-center text-text-muted-light dark:text-text-muted-dark hover:text-primary transition-colors no-underline"
          >
            <Icon name="arrow_back" class="text-xl" />
          </a>
          <h2 class="text-2xl font-bold">{t("channelDetail.playlists")}</h2>
        </div>

        {channelInfo && (
          <ChannelInfoCard
            channelInfo={channelInfo}
            isSubscribed={isSubscribed}
            onToggleSubscription={() => requireAuth(handleToggleSubscription)}
            isToggling={isToggling}
          />
        )}

        {playlists.length > 0 ? (
          <>
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
              {playlists.map((pl) => (
                <a
                  key={pl.playlist_id}
                  href={`/playlists/${pl.playlist_id}`}
                  class="group flex flex-col bg-card-light dark:bg-card-dark rounded-xl border border-transparent hover:border-primary/20 transition-all duration-300 overflow-hidden no-underline"
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
                        <Icon name="playlist_play" class="text-5xl text-text-muted-light dark:text-text-muted-dark" />
                      </div>
                    )}
                  </div>
                  <div class="p-5">
                    <h3 class="text-xl font-bold text-charcoal dark:text-white leading-tight group-hover:text-primary transition-colors">
                      {pl.playlist_title}
                    </h3>
                    <span class="text-xs text-text-muted-light dark:text-text-muted-dark mt-2 block">
                      {t("playlists.videoCount", { count: pl.playlist_video_count })}
                    </span>
                  </div>
                </a>
              ))}
            </div>
            <div ref={sentinelRef} class="h-1" />
            {isLoadingMore && <LoadingSpinner size="sm" className="py-8" />}
            {!hasNext && !isLoadingMore && playlists.length > 0 && (
              <p class="text-center text-sm text-text-muted-light dark:text-text-muted-dark py-8">
                {t("dashboard.endOfFeed")}
              </p>
            )}
          </>
        ) : (
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="playlist_play" class="text-5xl mb-4" />
            <p class="text-lg font-medium">{t("playlists.empty")}</p>
          </div>
        )}
      </div>
      <AuthPromptDialog open={showAuthPrompt} onClose={closeAuthPrompt} />
    </DashboardLayout>
  );
}

export default function ChannelPlaylists({ channelId }: { channelId: string }) {
  return <ChannelPlaylistsContent channelId={channelId} />;
}
