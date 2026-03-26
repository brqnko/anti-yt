import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { AddPlaylistDialog } from "../../components/AddPlaylistDialog";
import { getPlaylist } from "../../api/generated/playlist";
import { formatTimeAgo } from "../../utils/format";
import { PAGE_SIZES } from "../../constants";
import type { GetPlaylists200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";

function PlaylistCard({ playlist }: { playlist: GetPlaylists200ItemsItem }) {
  const { t } = useTranslation();
  return (
    <a
      href={`/playlists/${playlist.playlist_id}`}
      class="group relative flex flex-col bg-card-light dark:bg-card-dark rounded-xl hover:-translate-y-0.5 border border-transparent hover:border-primary/20 transition-all duration-300 overflow-hidden no-underline"
    >
      <div class="relative aspect-video w-full overflow-hidden bg-gray-100 dark:bg-gray-800">
        {playlist.top_video_thumbnail_url ? (
          <img
            src={playlist.top_video_thumbnail_url}
            alt={playlist.playlist_title}
            loading="lazy"
            class="absolute inset-0 w-full h-full object-cover"
          />
        ) : (
          <div class="absolute inset-0 flex items-center justify-center">
            <Icon name="playlist_play" class="text-5xl text-text-muted-light dark:text-text-muted-dark" />
          </div>
        )}
        <div class="absolute inset-0 bg-black/10 group-hover:bg-black/0 transition-colors" />
      </div>
      <div class="flex flex-col flex-1 p-5 gap-3">
        <h3 class="text-xl font-bold text-charcoal dark:text-white leading-tight group-hover:text-primary transition-colors">
          {playlist.playlist_title}
        </h3>
        {playlist.playlist_description && (
          <p class="text-text-muted-light dark:text-text-muted-dark text-sm line-clamp-2">
            {playlist.playlist_description}
          </p>
        )}
        <div class="flex items-center gap-3 mt-auto text-text-muted-light dark:text-text-muted-dark text-xs">
          <span>
            {t("playlists.videoCount", {
              count: playlist.playlist_video_count,
            })}
          </span>
          <span>
            {t("playlists.createdAt", {
              time: formatTimeAgo(playlist.playlist_registered_at, t),
            })}
          </span>
          <span>
            {t("playlists.lastUpdated", {
              time: formatTimeAgo(playlist.playlist_updated_at, t),
            })}
          </span>
        </div>
      </div>
    </a>
  );
}

function PlaylistsContent() {
  const { t } = useTranslation();
  const { route } = useLocation();
  useTitle(t("playlists.pageTitle"));

  const [playlists, setPlaylists] = useState<GetPlaylists200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [error, setError] = useState(false);
  const [showAddPlaylist, setShowAddPlaylist] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);
  const hasNextRef = useRef(false);
  const loadingMoreRef = useRef(false);

  const loadInitial = useCallback(async () => {
    setIsLoading(true);
    setError(false);
    try {
      const res = await getPlaylist().getPlaylists({ limit: PAGE_SIZES.PLAYLISTS });
      setPlaylists(res.items);
      setHasNext(res.has_next);
      hasNextRef.current = res.has_next;
      cursorRef.current = res.items[res.items.length - 1]?.playlist_id;
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
      const res = await getPlaylist().getPlaylists({
        limit: PAGE_SIZES.PLAYLISTS,
        cursor: cursorRef.current,
      });
      setPlaylists((prev) => [...prev, ...res.items]);
      setHasNext(res.has_next);
      hasNextRef.current = res.has_next;
      cursorRef.current = res.items[res.items.length - 1]?.playlist_id;
    } catch {
      // Stop further pagination attempts on error
      hasNextRef.current = false;
      setHasNext(false);
    } finally {
      loadingMoreRef.current = false;
      setIsLoadingMore(false);
    }
  }, []);

  const sentinelRef = useInfiniteScroll(loadMore);

  return (
    <DashboardLayout>
      <div class="w-full max-w-[1200px] mx-auto px-6 py-6 lg:py-10 flex flex-col gap-8">
        {/* Page Heading & Actions */}
        <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-6 pb-6 border-b border-border-light dark:border-border-dark">
          <div class="flex flex-col gap-2 max-w-2xl">
            <h1 class="text-4xl font-black tracking-tight text-charcoal dark:text-white">
              {t("playlists.title")}
            </h1>
          </div>
          <div class="flex flex-wrap items-center gap-3">
            <button
              class="flex items-center gap-1.5 h-9 px-4 rounded-lg bg-primary hover:bg-primary/90 text-white text-sm font-bold hover:-translate-y-px transition-all cursor-pointer border-none"
              onClick={() => setShowAddPlaylist(true)}
            >
              <Icon name="add" class="text-lg" />
              {t("playlists.createNew")}
            </button>
          </div>
        </div>

        {/* Content */}
        {isLoading ? (
          <LoadingSpinner />
        ) : error ? (
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="error_outline" class="text-5xl mb-4" />
            <p class="text-lg font-medium">{t("playlists.loadError")}</p>
            <button
              onClick={loadInitial}
              class="mt-4 text-sm text-primary hover:underline bg-transparent border-none cursor-pointer"
            >
              {t("playlists.retry")}
            </button>
          </div>
        ) : playlists.length > 0 ? (
          <>
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
              {playlists.map((playlist) => (
                <PlaylistCard
                  key={playlist.playlist_id}
                  playlist={playlist}
                />
              ))}
            </div>
            {hasNext && <div ref={sentinelRef} class="h-1" />}
            {isLoadingMore && <LoadingSpinner size="sm" className="py-8" />}
          </>
        ) : (
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="playlist_play" class="text-5xl mb-4" />
            <p class="text-lg font-medium">{t("playlists.empty")}</p>
          </div>
        )}
      </div>

      <AddPlaylistDialog
        open={showAddPlaylist}
        onClose={() => setShowAddPlaylist(false)}
        onAdded={(pl) => route(`/playlists/${pl.playlist_id}`)}
      />
    </DashboardLayout>
  );
}

export default function Playlists() {
  return (
    <ProtectedRoute>
      <PlaylistsContent />
    </ProtectedRoute>
  );
}
