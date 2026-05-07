import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useEscapeKey } from "../../hooks/useEscapeKey";
import { useInfiniteScroll } from "../../hooks/useInfiniteScroll";
import { useRequireAuth } from "../../hooks/useRequireAuth";
import { DashboardLayout } from "../../components/DashboardLayout";
import { AuthPromptDialog } from "../../components/AuthPromptDialog";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { VideoCard } from "../../components/VideoCard";
import { Dialog } from "../../components/Dialog";
import { getPlaylist } from "../../api/generated/playlist";
import { getApiErrorCode } from "../../utils/api-error";
import { formatTimeAgo } from "../../utils/format";
import { buildWatchUrl } from "../../utils/url";
import { PAGE_SIZES } from "../../constants";
import { Linkify } from "../../components/Linkify";
import { Icon } from "../../components/Icon";
import type {
  GetPlaylistsPlaylistId200,
  GetPlaylistsPlaylistIdVideos200ItemsItem,
} from "../../api/generated/antiYtApi.schemas";


function EditPlaylistDialog({
  open,
  playlist,
  onClose,
  onSaved,
}: {
  open: boolean;
  playlist: GetPlaylistsPlaylistId200;
  onClose: () => void;
  onSaved: (p: { playlist_title: string; playlist_description: string }) => void;
}) {
  const { t } = useTranslation();
  const [title, setTitle] = useState(playlist.playlist_title);
  const [description, setDescription] = useState(playlist.playlist_description);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setTitle(playlist.playlist_title);
      setDescription(playlist.playlist_description);
      setError(null);
    }
  }, [open, playlist]);

  const handleSave = async () => {
    if (!title.trim() || isSaving) return;
    setIsSaving(true);
    setError(null);
    try {
      await getPlaylist().patchPlaylistsPlaylistId(playlist.playlist_id, {
        playlist_title: title.trim(),
        playlist_description: description.trim(),
      });
      onSaved({
        playlist_title: title.trim(),
        playlist_description: description.trim(),
      });

      onClose();
    } catch (err) {
      const code = getApiErrorCode(err);
      setError(code ? t(`apiErrors.${code}`, t("apiErrors.fallback")) : t("playlistDetail.editDialog.error"));
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} ariaLabel={t("playlistDetail.editDialog.title")} panelClass="flex flex-col gap-4">
        <h2 class="text-xl font-bold text-charcoal dark:text-white">
          {t("playlistDetail.editDialog.title")}
        </h2>
        <div class="flex flex-col gap-3">
          <div>
            <label class="block text-sm font-medium text-charcoal dark:text-white mb-1">
              {t("playlistDetail.editDialog.titleLabel")}
            </label>
            <input
              type="text"
              class="w-full h-10 px-3 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none text-charcoal dark:text-white"
              value={title}
              onInput={(e) => setTitle((e.target as HTMLInputElement).value)}
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-charcoal dark:text-white mb-1">
              {t("playlistDetail.editDialog.descriptionLabel")}
            </label>
            <textarea
              class="w-full h-24 px-3 py-2 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none text-charcoal dark:text-white resize-none"
              value={description}
              onInput={(e) =>
                setDescription((e.target as HTMLTextAreaElement).value)
              }
            />
          </div>
        </div>
        {error && (
          <p class="text-sm text-red-500" role="alert">
            {error}
          </p>
        )}
        <div class="flex justify-end gap-3 pt-2">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer bg-transparent border-none"
            onClick={onClose}
          >
            {t("playlistDetail.editDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-bold bg-primary text-white hover:bg-primary/90 cursor-pointer border-none disabled:opacity-50"
            onClick={handleSave}
            disabled={!title.trim() || isSaving}
          >
            {isSaving
              ? t("playlistDetail.editDialog.saving")
              : t("playlistDetail.editDialog.save")}
          </button>
        </div>
    </Dialog>
  );
}

function DeleteConfirmDialog({
  open,
  onClose,
  onConfirm,
  isDeleting,
}: {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isDeleting: boolean;
}) {
  const { t } = useTranslation();

  return (
    <Dialog open={open} onClose={onClose} ariaLabel={t("playlistDetail.deleteDialog.title")} maxWidth="max-w-sm" panelClass="flex flex-col gap-4">
        <h2 class="text-xl font-bold text-charcoal dark:text-white">
          {t("playlistDetail.deleteDialog.title")}
        </h2>
        <p class="text-sm text-text-muted-light dark:text-text-muted-dark">
          {t("playlistDetail.deleteDialog.description")}
        </p>
        <div class="flex justify-end gap-3 pt-2">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer bg-transparent border-none"
            onClick={onClose}
          >
            {t("playlistDetail.deleteDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-bold bg-red-600 text-white hover:bg-red-700 cursor-pointer border-none disabled:opacity-50"
            onClick={onConfirm}
            disabled={isDeleting}
          >
            {isDeleting
              ? t("playlistDetail.deleteDialog.deleting")
              : t("playlistDetail.deleteDialog.confirm")}
          </button>
        </div>
    </Dialog>
  );
}

function RemoveVideoDialog({
  open,
  videoTitle,
  onClose,
  onConfirm,
  isRemoving,
  error,
}: {
  open: boolean;
  videoTitle: string;
  onClose: () => void;
  onConfirm: () => void;
  isRemoving: boolean;
  error: boolean;
}) {
  const { t } = useTranslation();

  return (
    <Dialog open={open} onClose={onClose} ariaLabel={t("playlistDetail.removeVideoDialog.title")} maxWidth="max-w-sm" panelClass="flex flex-col gap-4">
        <h2 class="text-xl font-bold text-charcoal dark:text-white">
          {t("playlistDetail.removeVideoDialog.title")}
        </h2>
        <p class="text-sm text-text-muted-light dark:text-text-muted-dark">
          {t("playlistDetail.removeVideoDialog.description", { title: videoTitle })}
        </p>
        {error && (
          <p class="text-sm text-red-500" role="alert">
            {t("playlistDetail.removeVideoDialog.error")}
          </p>
        )}
        <div class="flex justify-end gap-3 pt-2">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer bg-transparent border-none"
            onClick={onClose}
          >
            {t("playlistDetail.removeVideoDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-bold bg-red-600 text-white hover:bg-red-700 cursor-pointer border-none disabled:opacity-50"
            onClick={onConfirm}
            disabled={isRemoving}
          >
            {isRemoving
              ? t("playlistDetail.removeVideoDialog.removing")
              : t("playlistDetail.removeVideoDialog.confirm")}
          </button>
        </div>
    </Dialog>
  );
}

function AddVideoDialog({
  open,
  playlistId,
  onClose,
  onAdded,
}: {
  open: boolean;
  playlistId: string;
  onClose: () => void;
  onAdded: () => void;
}) {
  const { t } = useTranslation();
  const [text, setText] = useState("");
  const [isAdding, setIsAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setText("");
      setError(null);
    }
  }, [open]);

  const handleSubmit = async () => {
    const trimmed = text.trim();
    if (!trimmed || isAdding) return;
    setIsAdding(true);
    setError(null);
    try {
      await getPlaylist().postPlaylistsPlaylistIdVideos(playlistId, {
        external_video_text: trimmed,
      });
      onAdded();
      onClose();
    } catch (err) {
      const code = getApiErrorCode(err);
      setError(code ? t(`apiErrors.${code}`, t("playlistDetail.addVideoError")) : t("playlistDetail.addVideoError"));
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} ariaLabel={t("playlistDetail.addVideo")} panelClass="flex flex-col gap-4">
        <h2 class="text-xl font-bold text-charcoal dark:text-white">
          {t("playlistDetail.addVideo")}
        </h2>
        <div class="relative">
          <button
            type="button"
            class="absolute inset-y-0 left-0 flex items-center pl-3 pr-1 text-text-muted-light dark:text-text-muted-dark hover:text-primary bg-transparent border-none cursor-pointer"
            aria-label={t("playlistDetail.addVideoPaste")}
            onClick={async () => {
              try {
                const clipText = await navigator.clipboard.readText();
                if (clipText) setText(clipText);
              } catch {}
            }}
          >
            <Icon name="content_paste" class="text-[20px]" />
          </button>
          <input
            type="text"
            class="w-full pl-10 pr-4 py-3 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white placeholder-taupe focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all text-sm"
            placeholder={t("playlistDetail.addVideoPlaceholder")}
            value={text}
            onInput={(e) => setText((e.target as HTMLInputElement).value)}
            onKeyDown={(e) => { if (e.key === "Enter") handleSubmit(); }}
            disabled={isAdding}
          />
        </div>
        {error && (
          <p class="text-sm text-red-500" role="alert">{error}</p>
        )}
        <div class="flex justify-end gap-3 pt-2">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer bg-transparent border-none"
            onClick={onClose}
          >
            {t("playlistDetail.addVideoCancel")}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-bold bg-primary text-white hover:bg-primary/90 cursor-pointer border-none disabled:opacity-50"
            onClick={handleSubmit}
            disabled={!text.trim() || isAdding}
          >
            {isAdding ? t("playlistDetail.addVideoAdding") : t("playlistDetail.addVideoButton")}
          </button>
        </div>
    </Dialog>
  );
}

function CopyPlaylistDialog({
  open,
  playlist,
  onClose,
  onCopied,
}: {
  open: boolean;
  playlist: GetPlaylistsPlaylistId200;
  onClose: () => void;
  onCopied: (playlistId: string) => void;
}) {
  const { t } = useTranslation();
  const [title, setTitle] = useState(playlist.playlist_title);
  const [description, setDescription] = useState(playlist.playlist_description);
  const [isCopying, setIsCopying] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setTitle(playlist.playlist_title);
      setDescription(playlist.playlist_description);
      setError(null);
    }
  }, [open, playlist]);

  const handleCopy = async () => {
    if (!title.trim() || isCopying) return;
    setIsCopying(true);
    setError(null);
    try {
      const res = await getPlaylist().postPlaylistsPlaylistIdCopy(playlist.playlist_id, {
        playlist_title: title.trim(),
        playlist_description: description.trim(),
      });
      onCopied(res.playlist_id);
    } catch (err) {
      const code = getApiErrorCode(err);
      setError(code ? t(`apiErrors.${code}`, t("apiErrors.fallback")) : t("playlistDetail.copyDialog.error"));
    } finally {
      setIsCopying(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} ariaLabel={t("playlistDetail.copyDialog.title")} panelClass="flex flex-col gap-4">
        <h2 class="text-xl font-bold text-charcoal dark:text-white">
          {t("playlistDetail.copyDialog.title")}
        </h2>
        <div class="flex flex-col gap-3">
          <div>
            <label class="block text-sm font-medium text-charcoal dark:text-white mb-1">
              {t("playlistDetail.copyDialog.titleLabel")}
            </label>
            <input
              type="text"
              class="w-full h-10 px-3 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none text-charcoal dark:text-white"
              value={title}
              onInput={(e) => setTitle((e.target as HTMLInputElement).value)}
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-charcoal dark:text-white mb-1">
              {t("playlistDetail.copyDialog.descriptionLabel")}
            </label>
            <textarea
              class="w-full h-24 px-3 py-2 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none text-charcoal dark:text-white resize-none"
              value={description}
              onInput={(e) =>
                setDescription((e.target as HTMLTextAreaElement).value)
              }
            />
          </div>
        </div>
        {error && (
          <p class="text-sm text-red-500" role="alert">
            {error}
          </p>
        )}
        <div class="flex justify-end gap-3 pt-2">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer bg-transparent border-none"
            onClick={onClose}
          >
            {t("playlistDetail.copyDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-bold bg-primary text-white hover:bg-primary/90 cursor-pointer border-none disabled:opacity-50"
            onClick={handleCopy}
            disabled={!title.trim() || isCopying}
          >
            {isCopying
              ? t("playlistDetail.copyDialog.copying")
              : t("playlistDetail.copyDialog.copy")}
          </button>
        </div>
    </Dialog>
  );
}

function PlaylistDetailContent({ playlistId }: { playlistId: string }) {
  const { t } = useTranslation();
  const { route } = useLocation();
  const { requireAuth, showAuthPrompt, closeAuthPrompt } = useRequireAuth();

  const [playlistInfo, setPlaylistInfo] = useState<GetPlaylistsPlaylistId200 | null>(null);
  const [videos, setVideos] = useState<GetPlaylistsPlaylistIdVideos200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [error, setError] = useState(false);
  const [notFound, setNotFound] = useState(false);
  const cursorRef = useRef<string | undefined>(undefined);

  const [showEdit, setShowEdit] = useState(false);
  const [showDelete, setShowDelete] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [removeTarget, setRemoveTarget] = useState<GetPlaylistsPlaylistIdVideos200ItemsItem | null>(null);
  const [isRemoving, setIsRemoving] = useState(false);
  const [removeError, setRemoveError] = useState(false);

  const [showAddVideo, setShowAddVideo] = useState(false);
  const [showCopy, setShowCopy] = useState(false);

  useTitle(playlistInfo?.playlist_title ?? t("playlistDetail.pageTitle"));

  const loadInitial = useCallback(async () => {
    setIsLoading(true);
    setError(false);
    setNotFound(false);
    try {
      const [infoRes, videosRes] = await Promise.allSettled([
        getPlaylist().getPlaylistsPlaylistId(playlistId),
        getPlaylist().getPlaylistsPlaylistIdVideos(playlistId, { limit: PAGE_SIZES.PLAYLIST_VIDEOS }),
      ]);

      if (infoRes.status === "fulfilled") {
        setPlaylistInfo(infoRes.value);
      } else {
        const err = infoRes.reason;
        if (err?.response?.status === 404) {
          setNotFound(true);
        } else {
          setError(true);
        }
        return;
      }

      if (videosRes.status === "fulfilled") {
        setVideos(videosRes.value.items);
        setHasNext(videosRes.value.has_next);
        const lastItem =
          videosRes.value.items[videosRes.value.items.length - 1];
        cursorRef.current = lastItem?.video_id;
      }
    } catch {
      setError(true);
    } finally {
      setIsLoading(false);
    }
  }, [playlistId]);

  useEffect(() => {
    loadInitial();
  }, [loadInitial]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasNext) return;
    setIsLoadingMore(true);
    try {
      const res = await getPlaylist().getPlaylistsPlaylistIdVideos(playlistId, {
        limit: 20,
        cursor: cursorRef.current,
      });
      setVideos((prev) => [...prev, ...res.items]);
      setHasNext(res.has_next);
      const lastItem = res.items[res.items.length - 1];
      cursorRef.current = lastItem?.video_id;
    } catch {
      setHasNext(false);
    } finally {
      setIsLoadingMore(false);
    }
  }, [playlistId, isLoadingMore, hasNext]);

  const sentinelRef = useInfiniteScroll(loadMore);

  const handleDelete = async () => {
    if (isDeleting) return;
    setIsDeleting(true);
    try {
      await getPlaylist().deletePlaylistsPlaylistId(playlistId);

      route("/playlists");
    } catch {
      // keep dialog open on error
    } finally {
      setIsDeleting(false);
    }
  };

  const handleRemoveVideo = async () => {
    if (!removeTarget || isRemoving) return;
    setIsRemoving(true);
    setRemoveError(false);
    try {
      await getPlaylist().deletePlaylistsPlaylistIdVideos(playlistId, {
        video_id: removeTarget.video_id,
      });
      setVideos((prev) =>
        prev.filter((v) => v.video_id !== removeTarget.video_id),
      );
      const refreshed = await getPlaylist().getPlaylistsPlaylistId(playlistId);
      setPlaylistInfo(refreshed);
      setRemoveTarget(null);
    } catch {
      setRemoveError(true);
    } finally {
      setIsRemoving(false);
    }
  };

  if (isLoading) {
    return (
      <DashboardLayout>
        <LoadingSpinner className="py-32" />
      </DashboardLayout>
    );
  }

  if (notFound) {
    return (
      <DashboardLayout>
        <div class="w-full max-w-[1200px] mx-auto px-6 py-10">
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="playlist_remove" class="text-5xl mb-4" />
            <p class="text-lg font-medium">
              {t("playlistDetail.notFound")}
            </p>
<a
              class="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 no-underline"
              href="/playlists"
            >
              <Icon name="arrow_back" class="text-[18px]" />
              {t("playlistDetail.backToPlaylists")}
            </a>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  if (error || !playlistInfo) {
    return (
      <DashboardLayout>
        <div class="w-full max-w-[1200px] mx-auto px-6 py-10">
          <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
            <Icon name="error_outline" class="text-5xl mb-4" />
            <p class="text-lg font-medium">
              {t("playlistDetail.loadError")}
            </p>
            <button
              onClick={loadInitial}
              class="mt-4 text-sm text-primary hover:underline bg-transparent border-none cursor-pointer"
            >
              {t("playlistDetail.retry")}
            </button>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div class="flex-1 overflow-y-auto w-full max-w-[1200px] mx-auto px-6 py-6 lg:py-10">
        {/* Back link */}
        <a
          href="/playlists"
          class="inline-flex items-center gap-1 text-sm text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white no-underline mb-6"
        >
          <Icon name="arrow_back" class="text-[18px]" />
          {t("playlistDetail.backToPlaylists")}
        </a>

        {/* Playlist Header */}
        <div class="bg-card-light dark:bg-card-dark rounded-xl border border-border-light dark:border-border-dark mb-8 p-6">
          <div class="flex flex-col sm:flex-row gap-6 items-start">
            {/* Thumbnail */}
            {videos.length > 0 ? (
              <a
                href={buildWatchUrl(videos[0].video_id, undefined, playlistId)}
                class="group/thumb relative w-full sm:w-48 aspect-video flex-shrink-0 rounded-lg overflow-hidden bg-gray-100 dark:bg-gray-800 block no-underline"
              >
                {playlistInfo.top_video_thumbnail_url ? (
                  <img
                    src={playlistInfo.top_video_thumbnail_url}
                    alt={playlistInfo.playlist_title}
                    class="absolute inset-0 w-full h-full object-cover"
                  />
                ) : (
                  <div class="absolute inset-0 flex items-center justify-center">
                    <Icon name="playlist_play" class="text-5xl text-text-muted-light dark:text-text-muted-dark" />
                  </div>
                )}
                <div class="absolute inset-0 bg-black/30 opacity-0 group-hover/thumb:opacity-100" />
                <div class="absolute inset-0 flex items-center justify-center opacity-0 group-hover/thumb:opacity-100 pointer-events-none">
                  <div class="size-12 rounded-full bg-primary/90 flex items-center justify-center text-white">
                    <Icon name="play_arrow" class="text-[28px] ml-1" />
                  </div>
                </div>
              </a>
            ) : (
              <div class="relative w-full sm:w-48 aspect-video flex-shrink-0 rounded-lg overflow-hidden bg-gray-100 dark:bg-gray-800">
                <div class="absolute inset-0 flex items-center justify-center">
                  <Icon name="playlist_play" class="text-5xl text-text-muted-light dark:text-text-muted-dark" />
                </div>
              </div>
            )}

            {/* Info */}
            <div class="flex-1 min-w-0">
              <h1 class="text-2xl md:text-3xl font-bold text-charcoal dark:text-white mb-2">
                {playlistInfo.playlist_title}
              </h1>
              {playlistInfo.playlist_description && (
                <p class="text-text-muted-light dark:text-text-muted-dark text-sm mb-3 whitespace-pre-wrap break-words">
                  <Linkify text={playlistInfo.playlist_description} />
                </p>
              )}
              <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-text-muted-light dark:text-text-muted-dark">
                <span>
                  {t("playlists.videoCount", {
                    count: playlistInfo.playlist_video_count,
                  })}
                </span>
                <span>
                  {t("playlists.createdAt", {
                    time: formatTimeAgo(playlistInfo.playlist_registered_at, t),
                  })}
                </span>
                <span>
                  {t("playlists.lastUpdated", {
                    time: formatTimeAgo(playlistInfo.playlist_updated_at, t),
                  })}
                </span>
              </div>
            </div>

            {/* Actions */}
            {playlistInfo.playlist_type === "normal" ? (
              <div class="flex gap-2 flex-shrink-0">
                <button
                  class="flex items-center gap-1.5 h-9 px-3 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary/90 cursor-pointer border-none"
                  onClick={() => requireAuth(() => setShowAddVideo(true))}
                >
                  <Icon name="add" class="text-[18px]" />
                  {t("playlistDetail.addVideo")}
                </button>
                <button
                  class="flex items-center gap-1.5 h-9 px-3 rounded-lg bg-transparent border border-border-light dark:border-border-dark text-sm font-medium text-charcoal dark:text-white hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer"
                  onClick={() => requireAuth(() => setShowEdit(true))}
                >
                  <Icon name="edit" class="text-[18px]" />
                  {t("playlistDetail.edit")}
                </button>
                <button
                  class="flex items-center gap-1.5 h-9 px-3 rounded-lg bg-transparent border border-red-300 dark:border-red-800 text-sm font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 cursor-pointer"
                  onClick={() => requireAuth(() => setShowDelete(true))}
                >
                  <Icon name="delete" class="text-[18px]" />
                  {t("playlistDetail.delete")}
                </button>
              </div>
            ) : (
              <div class="flex gap-2 flex-shrink-0">
                <button
                  class="flex items-center gap-1.5 h-9 px-3 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary/90 cursor-pointer border-none"
                  onClick={() => requireAuth(() => setShowCopy(true))}
                >
                  <Icon name="content_copy" class="text-[18px]" />
                  {t("playlistDetail.copy")}
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Videos */}
        <div>
          <h3 class="text-lg font-bold text-charcoal dark:text-white mb-4">
            {t("playlistDetail.videos")}
          </h3>

          {videos.length > 0 ? (
            <>
              <div class="flex flex-col divide-y divide-gray-200 dark:divide-gray-800">
                {videos.map((video) => (
                  <div
                    key={video.video_id}
                    class="py-4 first:pt-0 group/row flex items-start sm:items-center gap-2"
                  >
                    <div class="flex-1 min-w-0">
                      <VideoCard
                        layout="row"
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
                        playlistId={playlistId}
                      />
                    </div>
                    {playlistInfo.playlist_type === "normal" && (
                      <button
                        class="flex-shrink-0 p-1.5 rounded-lg text-text-muted-light dark:text-text-muted-dark hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 cursor-pointer bg-transparent border-none opacity-100 lg:opacity-0 lg:group-hover/row:opacity-100 focus:opacity-100"
                        onClick={() => setRemoveTarget(video)}
                        title={t("playlistDetail.removeVideo")}
                      >
                        <Icon name="close" class="text-[20px]" />
                      </button>
                    )}
                  </div>
                ))}
              </div>
              <div ref={sentinelRef} class="h-1" />
              {isLoadingMore && (
                <LoadingSpinner size="sm" className="py-8" />
              )}
              {!hasNext && !isLoadingMore && videos.length > 0 && (
                <p class="text-center text-sm text-text-muted-light dark:text-text-muted-dark py-8">
                  {t("playlistDetail.endOfList")}
                </p>
              )}
            </>
          ) : (
            <div class="flex flex-col items-center justify-center py-12 text-text-muted-light dark:text-text-muted-dark bg-card-light dark:bg-card-dark rounded-xl border border-border-light dark:border-border-dark">
              <Icon name="playlist_play" class="text-4xl mb-3" />
              <p class="text-sm font-medium">
                {t("playlistDetail.noVideos")}
              </p>
            </div>
          )}
        </div>
      </div>

      <EditPlaylistDialog
        open={showEdit}
        playlist={playlistInfo}
        onClose={() => setShowEdit(false)}
        onSaved={(updated) =>
          setPlaylistInfo((prev) => (prev ? { ...prev, ...updated } : prev))
        }
      />

      <DeleteConfirmDialog
        open={showDelete}
        onClose={() => setShowDelete(false)}
        onConfirm={handleDelete}
        isDeleting={isDeleting}
      />

      <RemoveVideoDialog
        open={!!removeTarget}
        videoTitle={removeTarget?.external_video_title ?? ""}
        onClose={() => {
          setRemoveTarget(null);
          setRemoveError(false);
        }}
        onConfirm={handleRemoveVideo}
        isRemoving={isRemoving}
        error={removeError}
      />

      <AddVideoDialog
        open={showAddVideo}
        playlistId={playlistId}
        onClose={() => setShowAddVideo(false)}
        onAdded={loadInitial}
      />

      {playlistInfo.playlist_type !== "normal" && (
        <CopyPlaylistDialog
          open={showCopy}
          playlist={playlistInfo}
          onClose={() => setShowCopy(false)}
          onCopied={(newPlaylistId) => {
            setShowCopy(false);
            route(`/playlists/${newPlaylistId}`);
          }}
        />
      )}
      <AuthPromptDialog open={showAuthPrompt} onClose={closeAuthPrompt} />
    </DashboardLayout>
  );
}

export default function PlaylistDetail({
  playlistId,
}: {
  playlistId: string;
}) {
  return <PlaylistDetailContent playlistId={playlistId} />;
}
