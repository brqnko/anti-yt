import { useState, useEffect, useRef, useCallback } from "preact/hooks";
import { useRoute, useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useRequireAuth } from "../../hooks/useRequireAuth";
import { DashboardLayout } from "../../components/DashboardLayout";
import { AuthPromptDialog } from "../../components/AuthPromptDialog";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { getVideo } from "../../api/generated/video";
import { getHistory } from "../../api/generated/history";
import { getPlaylist } from "../../api/generated/playlist";
import { Dialog } from "../../components/Dialog";
import { formatDuration, formatSubscriberCount, formatTimeAgo } from "../../utils/format";
import { buildWatchUrl } from "../../utils/url";
import { getApiErrorCode } from "../../utils/api-error";
import { PAGE_SIZES } from "../../constants";
import type {
  GetVideosVideoId200,
  GetPlaylists200ItemsItem,
  GetPlaylistsPlaylistId200,
  GetPlaylistsPlaylistIdVideos200ItemsItem,
} from "../../api/generated/antiYtApi.schemas";
import { useYouTubePlayer, PlayerState } from "./useYouTubePlayer";
import { useHeartbeat } from "./useHeartbeat";
import { Linkify } from "../../components/Linkify";
import { Icon } from "../../components/Icon";

const PLAYER_CONTAINER_ID = "yt-player";

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
            class="absolute inset-y-0 left-0 flex items-center pl-3 pr-1 text-text-muted-light dark:text-text-muted-dark hover:text-primary transition-colors bg-transparent border-none cursor-pointer"
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
            class="px-4 py-2 rounded-lg text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer bg-transparent border-none"
            onClick={onClose}
          >
            {t("playlistDetail.addVideoCancel")}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-bold bg-primary text-white hover:bg-primary/90 transition-colors cursor-pointer border-none disabled:opacity-50"
            onClick={handleSubmit}
            disabled={!text.trim() || isAdding}
          >
            {isAdding ? t("playlistDetail.addVideoAdding") : t("playlistDetail.addVideoButton")}
          </button>
        </div>
    </Dialog>
  );
}

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
            class="px-4 py-2 rounded-lg text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer bg-transparent border-none"
            onClick={onClose}
          >
            {t("playlistDetail.editDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-bold bg-primary text-white hover:bg-primary/90 transition-colors cursor-pointer border-none disabled:opacity-50"
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

function VideoPlayerContent() {
  const { t } = useTranslation();
  const { params } = useRoute();
  const { route } = useLocation();
  const { isAuthenticated, requireAuth, showAuthPrompt, closeAuthPrompt } = useRequireAuth();
  const videoId = params.videoId;

  const [video, setVideo] = useState<GetVideosVideoId200 | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(false);
  const startTimeRef = useRef<number | null>(
    typeof window !== "undefined"
      ? (() => {
          const t = new URLSearchParams(window.location.search).get("t");
          const n = Number(t);
          return t && Number.isFinite(n) && n > 0 ? Math.floor(n) : null;
        })()
      : null,
  );
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isDescExpanded, setIsDescExpanded] = useState(false);
  const [controlsVisible, setControlsVisible] = useState(false);
  const [isSeeking, setIsSeeking] = useState(false);

  const [playerHeight, setPlayerHeight] = useState<number | null>(null);
  const playerWrapperRef = useRef<HTMLDivElement>(null);
  const progressBarRef = useRef<HTMLDivElement>(null);
  const progressFillRef = useRef<HTMLDivElement>(null);
  const progressKnobRef = useRef<HTMLDivElement>(null);
  const isSeekingRef = useRef(false);
  const controlsTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const descRef = useRef<HTMLDivElement>(null);
  const [descOverflows, setDescOverflows] = useState(false);
  const lastPointerTypeRef = useRef<string>("mouse");

  // Playlist sidebar state
  const [playlistId] = useState<string | null>(() => {
    if (typeof window === "undefined") return null;
    return new URLSearchParams(window.location.search).get("playlist");
  });
  const [playlistInfo, setPlaylistInfo] = useState<GetPlaylistsPlaylistId200 | null>(null);
  const [playlistVideos, setPlaylistVideos] = useState<GetPlaylistsPlaylistIdVideos200ItemsItem[]>([]);
  const [playlistLoading, setPlaylistLoading] = useState(false);
  const [playlistHasNext, setPlaylistHasNext] = useState(false);
  const [playlistLoadingMore, setPlaylistLoadingMore] = useState(false);
  const playlistCursorRef = useRef<string | undefined>(undefined);
  const playlistLoadingMoreRef = useRef(false);
  const playlistHasNextRef = useRef(false);
  const [removingVideoId, setRemovingVideoId] = useState<string | null>(null);
  const [userPlaylists, setUserPlaylists] = useState<GetPlaylists200ItemsItem[]>([]);
  const [playlistDialogLoading, setPlaylistDialogLoading] = useState(false);
  const [addingToPlaylist, setAddingToPlaylist] = useState<string | null>(null);
  const [addedToPlaylist, setAddedToPlaylist] = useState<string | null>(null);
  const [failedToAddPlaylist, setFailedToAddPlaylist] = useState<string | null>(null);
  const [addedPlaylistIds, setAddedPlaylistIds] = useState<Set<string>>(new Set());
  const [showPlaylistDialog, setShowPlaylistDialog] = useState(false);
  const [showImportVideo, setShowImportVideo] = useState(false);
  const [showEditPlaylist, setShowEditPlaylist] = useState(false);
  const [markingWatched, setMarkingWatched] = useState(false);
  const [markedWatched, setMarkedWatched] = useState(false);
  const [markingWatchLater, setMarkingWatchLater] = useState(false);
  const [markedWatchLater, setMarkedWatchLater] = useState(false);

  // Refs for values used in keyboard handler / cleanup to avoid stale closures
  const durationRef = useRef(0);
  const volumeRef = useRef(100);
  const heartbeatTickRef = useRef<(() => void) | null>(null);

  // Refs for auto-play next video in playlist
  const playlistVideosRef = useRef(playlistVideos);
  playlistVideosRef.current = playlistVideos;
  const playlistIdRef = useRef(playlistId);
  playlistIdRef.current = playlistId;

  useTitle(video?.external_video_title ?? t("videoPlayer.pageTitle"));

  useEffect(() => {
    const el = playerWrapperRef.current;
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      setPlayerHeight(entries[0].contentRect.height);
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  useEffect(() => {
    if (!videoId) return;
    setIsLoading(true);
    setError(false);
    setMarkedWatched(false);
    setMarkedWatchLater(false);
    getVideo()
      .getVideosVideoId(videoId)
      .then((res) => {
        setVideo(res);
        setMarkedWatched(res.is_watched);
        setMarkedWatchLater(res.is_in_watch_later);
      })
      .catch(() => setError(true))
      .finally(() => setIsLoading(false));
  }, [videoId]);

  // Fetch playlist data when in playlist context
  useEffect(() => {
    if (!playlistId) return;
    setPlaylistLoading(true);
    Promise.allSettled([
      getPlaylist().getPlaylistsPlaylistId(playlistId),
      getPlaylist().getPlaylistsPlaylistIdVideos(playlistId, { limit: PAGE_SIZES.PLAYLIST_VIDEOS }),
    ]).then(([infoRes, videosRes]) => {
      if (infoRes.status === "fulfilled") {
        setPlaylistInfo(infoRes.value);
      }
      if (videosRes.status === "fulfilled") {
        setPlaylistVideos(videosRes.value.items);
        setPlaylistHasNext(videosRes.value.has_next);
        playlistHasNextRef.current = videosRes.value.has_next;
        const lastItem = videosRes.value.items[videosRes.value.items.length - 1];
        playlistCursorRef.current = lastItem?.video_id;
      }
    }).finally(() => setPlaylistLoading(false));
  }, [playlistId]);

  const loadMorePlaylistVideos = useCallback(async () => {
    if (playlistLoadingMoreRef.current || !playlistHasNextRef.current || !playlistId) return;
    playlistLoadingMoreRef.current = true;
    setPlaylistLoadingMore(true);
    try {
      const res = await getPlaylist().getPlaylistsPlaylistIdVideos(playlistId, {
        limit: PAGE_SIZES.PLAYLIST_VIDEOS,
        cursor: playlistCursorRef.current,
      });
      setPlaylistVideos((prev) => [...prev, ...res.items]);
      setPlaylistHasNext(res.has_next);
      playlistHasNextRef.current = res.has_next;
      const lastItem = res.items[res.items.length - 1];
      playlistCursorRef.current = lastItem?.video_id;
    } catch {
      playlistHasNextRef.current = false;
      setPlaylistHasNext(false);
    } finally {
      playlistLoadingMoreRef.current = false;
      setPlaylistLoadingMore(false);
    }
  }, [playlistId]);

  const handleRemoveFromPlaylist = useCallback(
    async (videoIdToRemove: string) => {
      if (removingVideoId || !playlistId) return;
      setRemovingVideoId(videoIdToRemove);
      try {
        await getPlaylist().deletePlaylistsPlaylistIdVideos(playlistId, {
          video_id: videoIdToRemove,
        });
        setPlaylistVideos((prev) => prev.filter((v) => v.video_id !== videoIdToRemove));
        setPlaylistInfo((prev) =>
          prev ? { ...prev, playlist_video_count: Math.max(0, prev.playlist_video_count - 1) } : prev,
        );
      } catch {
        // silently fail
      } finally {
        setRemovingVideoId(null);
      }
    },
    [playlistId, removingVideoId],
  );

  const openPlaylistDialog = useCallback(async () => {
    setShowPlaylistDialog(true);
    setPlaylistDialogLoading(true);
    try {
      const res = await getPlaylist().getPlaylists({ limit: 50 });
      setUserPlaylists(res.items);
    } catch {
      // silently fail
    } finally {
      setPlaylistDialogLoading(false);
    }
  }, []);

  const handleAddToPlaylist = useCallback(
    async (plId: string) => {
      if (addingToPlaylist || !videoId || addedPlaylistIds.has(plId)) return;
      setAddingToPlaylist(plId);
      setFailedToAddPlaylist(null);
      try {
        await getPlaylist().postPlaylistsPlaylistIdVideos(plId, {
          video_id: videoId,
        });
        setAddedToPlaylist(plId);
        setAddedPlaylistIds((prev) => new Set(prev).add(plId));
        setTimeout(() => setAddedToPlaylist(null), 2000);
      } catch {
        setFailedToAddPlaylist(plId);
        setTimeout(() => setFailedToAddPlaylist(null), 2000);
      } finally {
        setAddingToPlaylist(null);
      }
    },
    [addingToPlaylist, videoId],
  );

  const reloadPlaylist = useCallback(async () => {
    if (!playlistId) return;
    const [infoRes, videosRes] = await Promise.allSettled([
      getPlaylist().getPlaylistsPlaylistId(playlistId),
      getPlaylist().getPlaylistsPlaylistIdVideos(playlistId, { limit: PAGE_SIZES.PLAYLIST_VIDEOS }),
    ]);
    if (infoRes.status === "fulfilled") setPlaylistInfo(infoRes.value);
    if (videosRes.status === "fulfilled") {
      setPlaylistVideos(videosRes.value.items);
      setPlaylistHasNext(videosRes.value.has_next);
      playlistHasNextRef.current = videosRes.value.has_next;
      const lastItem = videosRes.value.items[videosRes.value.items.length - 1];
      playlistCursorRef.current = lastItem?.video_id;
    }
  }, [playlistId]);

  // Navigate to the next video in the playlist when the current video ends
  const handlePlayerStateChange = useCallback(
    (state: number) => {
      if (state !== PlayerState.ENDED) return;
      const plId = playlistIdRef.current;
      const videos = playlistVideosRef.current;
      if (!plId || videos.length === 0) return;

      const currentIdx = videos.findIndex((v) => v.video_id === videoId);
      if (currentIdx === -1 || currentIdx >= videos.length - 1) return;

      const next = videos[currentIdx + 1];
      route(buildWatchUrl(next.video_id, undefined, plId));
    },
    [videoId, route],
  );

  const {
    isReady,
    playerState,
    currentTime,
    currentTimeRef,
    duration,
    volume,
    isMuted,
    togglePlay,
    seekTo,
    setVolume,
    toggleMute,
    isLooping,
    toggleLoop,
    setHighFreqSync,
  } = useYouTubePlayer({
    videoId: video?.external_video_id ?? "",
    containerId: PLAYER_CONTAINER_ID,
    onStateChange: handlePlayerStateChange,
    onSyncTick: () => heartbeatTickRef.current?.(),
  });

  // Seek to start time from ?t= query param once player is ready
  useEffect(() => {
    if (isReady && startTimeRef.current != null) {
      seekTo(startTimeRef.current);
      startTimeRef.current = null;
    }
  }, [isReady, seekTo]);

  // Switch to high-frequency (rAF) time sync while controls are visible
  useEffect(() => {
    if (isReady && controlsVisible && playerState === PlayerState.PLAYING) {
      setHighFreqSync(true);
      return () => setHighFreqSync(false);
    }
  }, [isReady, controlsVisible, playerState, setHighFreqSync]);

  // Sync progress bar from the shared currentTimeRef (updated by useYouTubePlayer's rAF).
  // Uses a separate rAF only while the controls overlay is visible to avoid GPU work
  // compositing invisible layers every frame.
  useEffect(() => {
    if (!isReady || !duration || !controlsVisible) return;
    let raf: number;
    const tick = () => {
      if (!isSeekingRef.current) {
        const progress = Math.min(currentTimeRef.current / duration, 1);
        if (progressFillRef.current) {
          progressFillRef.current.style.transform = `scaleX(${progress})`;
        }
        if (progressKnobRef.current) {
          progressKnobRef.current.style.left = `${progress * 100}%`;
        }
      }
      raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [isReady, duration, currentTimeRef, controlsVisible]);

  // Keep refs in sync with latest values
  durationRef.current = duration;
  volumeRef.current = volume;
  isSeekingRef.current = isSeeking;

  const { remainingSeconds, tick: heartbeatTick } = useHeartbeat({
    videoId: isAuthenticated ? (video?.video_id ?? null) : null,
    playlistId,
    playerState,
    currentTimeRef,
    togglePlay,
  });
  heartbeatTickRef.current = heartbeatTick;

  // Check if description overflows (ResizeObserver for font-load safety)
  useEffect(() => {
    const el = descRef.current;
    if (!el) return;
    const observer = new ResizeObserver(() => {
      setDescOverflows(el.scrollHeight > el.clientHeight + 1);
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [video?.external_video_description]);

  // Fullscreen change listener
  useEffect(() => {
    const handler = () => {
      const fs = !!document.fullscreenElement;
      setIsFullscreen(fs);
      if (!fs) {
        screen.orientation?.unlock?.();
      }
    };
    document.addEventListener("fullscreenchange", handler);
    return () => document.removeEventListener("fullscreenchange", handler);
  }, []);

  const toggleFullscreen = useCallback(async () => {
    const el = playerWrapperRef.current;
    if (!el) return;
    if (document.fullscreenElement) {
      await document.exitFullscreen();
    } else {
      await el.requestFullscreen?.();
      try {
        await (screen.orientation as ScreenOrientation & { lock?: (orientation: string) => Promise<void> }).lock?.("landscape");
      } catch {
        // Screen Orientation API not supported or permission denied — ignore
      }
    }
  }, []);

  // Keyboard shortcuts (using refs to avoid re-registration every 250ms)
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (!isReady) return;
      if (e.ctrlKey || e.metaKey || e.altKey) return;

      switch (e.key) {
        case " ":
        case "k":
          e.preventDefault();
          togglePlay();
          break;
        case "ArrowLeft":
          e.preventDefault();
          seekTo(Math.max(0, currentTimeRef.current - 5));
          break;
        case "ArrowRight":
          e.preventDefault();
          seekTo(Math.min(durationRef.current, currentTimeRef.current + 5));
          break;
        case "ArrowUp":
          e.preventDefault();
          setVolume(Math.min(100, volumeRef.current + 5));
          break;
        case "ArrowDown":
          e.preventDefault();
          setVolume(Math.max(0, volumeRef.current - 5));
          break;
        case "m":
          e.preventDefault();
          toggleMute();
          break;
        case "f":
          e.preventDefault();
          toggleFullscreen();
          break;
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [isReady, togglePlay, seekTo, setVolume, toggleMute, toggleFullscreen]);

  // Show controls on touch/interaction, auto-hide after 3s
  const showControlsTemporarily = useCallback(() => {
    setControlsVisible(true);
    if (controlsTimerRef.current) clearTimeout(controlsTimerRef.current);
    controlsTimerRef.current = setTimeout(() => setControlsVisible(false), 3000);
  }, []);

  // Track pointer type for touch vs mouse distinction
  const handlePlayerAreaPointerDown = useCallback(
    (e: PointerEvent) => {
      lastPointerTypeRef.current = e.pointerType;
    },
    [],
  );

  const handlePlayerAreaClick = useCallback(() => {
    if (lastPointerTypeRef.current === "touch") {
      if (!controlsVisible) {
        showControlsTemporarily();
        return;
      }
    }
    togglePlay();
  }, [controlsVisible, showControlsTemporarily, togglePlay]);

  // Progress bar: click + drag support
  const calcSeekRatio = useCallback(
    (clientX: number) => {
      const bar = progressBarRef.current;
      if (!bar) return null;
      const rect = bar.getBoundingClientRect();
      return Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
    },
    [],
  );

  const handleProgressPointerDown = useCallback(
    (e: PointerEvent) => {
      if (!duration) return;
      e.preventDefault();
      (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
      setIsSeeking(true);
      const ratio = calcSeekRatio(e.clientX);
      if (ratio !== null) {
        if (progressFillRef.current) progressFillRef.current.style.transform = `scaleX(${ratio})`;
        if (progressKnobRef.current) progressKnobRef.current.style.left = `${ratio * 100}%`;
      }
    },
    [duration, calcSeekRatio],
  );

  const handleProgressPointerMove = useCallback(
    (e: PointerEvent) => {
      if (!isSeeking || !duration) return;
      const ratio = calcSeekRatio(e.clientX);
      if (ratio !== null) {
        if (progressFillRef.current) progressFillRef.current.style.transform = `scaleX(${ratio})`;
        if (progressKnobRef.current) progressKnobRef.current.style.left = `${ratio * 100}%`;
      }
    },
    [isSeeking, duration, calcSeekRatio],
  );

  const handleProgressPointerUp = useCallback(
    (e: PointerEvent) => {
      if (!isSeeking || !duration) return;
      (e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId);
      setIsSeeking(false);
      const ratio = calcSeekRatio(e.clientX);
      if (ratio !== null) {
        seekTo(ratio * duration);
      }
    },
    [isSeeking, duration, calcSeekRatio, seekTo],
  );

  const isPlaying = playerState === PlayerState.PLAYING;

  if (isLoading && !video) {
    return (
      <DashboardLayout>
        <LoadingSpinner className="flex-1" />
      </DashboardLayout>
    );
  }

  if (error || !video) {
    return (
      <DashboardLayout>
        <div class="flex flex-col items-center justify-center flex-1 text-text-muted-light dark:text-text-muted-dark">
          <Icon name="error" class="text-5xl mb-4" />
          <p class="text-lg font-medium">{t("videoPlayer.notFound")}</p>
          <a
            href="/"
            class="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 transition-colors no-underline"
          >
            {t("channelDetail.backToDashboard")}
          </a>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div class="flex-1 overflow-y-auto">
        <div class="max-w-[1536px] mx-auto px-0 sm:px-6 py-0 sm:py-8 pb-8 flex flex-col xl:flex-row xl:items-start gap-8">
          {/* Main content */}
          <div class="flex-1 min-w-0">
            {/* YouTube Player */}
            <div
              ref={playerWrapperRef}
              class="w-full bg-black overflow-hidden ring-1 ring-white/10 relative aspect-video group/player"
            >
              {/* YouTube iframe gets injected here */}
              <div id={PLAYER_CONTAINER_ID} class="absolute inset-0 w-full h-full" />

              {/* Click/tap overlay to toggle play/pause (above iframe) */}
              {isReady && (
                <div
                  class="absolute inset-0 z-10 cursor-pointer"
                  onPointerDown={handlePlayerAreaPointerDown}
                  onClick={handlePlayerAreaClick}
                  onMouseMove={showControlsTemporarily}
                />
              )}

              {/* Big play button when not started or paused */}
              {isReady && !isPlaying && playerState !== PlayerState.BUFFERING && (
                <div class="absolute inset-0 z-20 flex items-center justify-center pointer-events-none">
                  <button
                    class="size-20 rounded-full bg-primary text-white flex items-center justify-center border-none cursor-pointer pointer-events-auto"
                    onClick={togglePlay}
                    aria-label={t("videoPlayer.play")}
                  >
                    <Icon name="play_arrow" class="text-5xl" />
                  </button>
                </div>
              )}




              {/* Player controls overlay */}
              <div
                class={`absolute bottom-0 inset-x-0 p-4 md:p-6 bg-gradient-to-t from-black/80 to-transparent z-30 transition-[opacity,visibility] ${controlsVisible ? "opacity-100 visible" : "opacity-0 invisible group-hover/player:visible group-hover/player:opacity-100"}`}
                onMouseMove={showControlsTemporarily}
              >
                {/* Progress bar */}
                <div class="flex items-center gap-4 mb-3">
                  <div
                    ref={progressBarRef}
                    role="slider"
                    aria-label={t("videoPlayer.seekBar")}
                    aria-valuemin={0}
                    aria-valuemax={Math.floor(duration)}
                    aria-valuenow={Math.floor(currentTime)}
                    aria-valuetext={`${formatDuration(currentTime)} / ${formatDuration(duration)}`}
                    tabIndex={0}
                    class={`flex-1 h-1.5 bg-white/20 rounded-full relative cursor-pointer group/progress touch-none ${isSeeking ? "h-2.5" : ""}`}
                    onPointerDown={handleProgressPointerDown}
                    onPointerMove={handleProgressPointerMove}
                    onPointerUp={handleProgressPointerUp}
                  >
                    <div
                      ref={progressFillRef}
                      class="absolute inset-y-0 left-0 w-full bg-primary rounded-full origin-left"
                      style={{ transform: `scaleX(${duration > 0 ? Math.min(currentTime / duration, 1) : 0})` }}
                    />
                    <div
                      ref={progressKnobRef}
                      class={`absolute top-1/2 size-4 bg-primary border-2 border-white rounded-full transition-transform -translate-x-1/2 -translate-y-1/2 ${isSeeking ? "scale-100" : "scale-0 group-hover/progress:scale-100"}`}
                      style={{ left: `${duration > 0 ? Math.min(currentTime / duration, 1) * 100 : 0}%` }}
                    />
                  </div>
                </div>
                {/* Controls row */}
                <div class="flex items-center justify-between text-white text-sm">
                  <div class="flex items-center gap-4 md:gap-6">
                    <button
                      class="bg-transparent border-none p-0 cursor-pointer text-white"
                      onClick={togglePlay}
                      aria-label={isPlaying ? t("videoPlayer.pause") : t("videoPlayer.play")}
                    >
                      <Icon name={isPlaying ? "pause" : "play_arrow"} />
                    </button>
                    <button
                      class="bg-transparent border-none p-0 cursor-pointer text-white"
                      onClick={toggleMute}
                      aria-label={isMuted ? t("videoPlayer.unmute") : t("videoPlayer.mute")}
                    >
                      <Icon name={isMuted || volume === 0
                          ? "volume_off"
                          : volume < 50
                            ? "volume_down"
                            : "volume_up"} />
                    </button>
                    <input
                      type="range"
                      min="0"
                      max="100"
                      value={isMuted ? 0 : volume}
                      onInput={(e) => setVolume(Number((e.target as HTMLInputElement).value))}
                      class="w-20 h-1 accent-primary cursor-pointer hidden sm:block"
                      aria-label={t("videoPlayer.volume")}
                    />
                    <span class="font-mono text-xs opacity-80">
                      {formatDuration(currentTime)} / {formatDuration(duration)}
                    </span>
                  </div>
                  <div class="flex items-center gap-4 md:gap-6">
                    {remainingSeconds !== null && (
                      <span class="text-xs opacity-70 hidden sm:inline">
                        {t("videoPlayer.remaining")}: {formatDuration(remainingSeconds)}
                      </span>
                    )}
                    <button
                      class="bg-transparent border-none p-0 cursor-pointer text-white"
                      onClick={toggleFullscreen}
                      aria-label={isFullscreen ? t("videoPlayer.exitFullscreen") : t("videoPlayer.fullscreen")}
                    >
                      <Icon name={isFullscreen ? "fullscreen_exit" : "fullscreen"} />
                    </button>
                  </div>
                </div>
              </div>
            </div>

            {/* Video info */}
            <div class="mt-8 px-4 sm:px-0">
              <h1 class="text-xl font-bold leading-tight tracking-tight">
                {video.external_video_title}
              </h1>
              <div class="mt-4 pb-6 border-b border-border-light dark:border-border-dark">
                <div class="flex items-center gap-4">
                    <a
                      href={`/channels/${video.channel_id}`}
                      class="size-12 rounded-full bg-cover bg-center border-2 border-primary/20 overflow-hidden block flex-shrink-0"
                    >
                      <img
                        src={video.external_channel_icon_url}
                        alt={video.external_channel_display_name}
                        loading="lazy"
                        class="w-full h-full object-cover"
                      />
                    </a>
                    <div>
                      <a
                        href={`/channels/${video.channel_id}`}
                        class="font-bold text-lg no-underline text-charcoal dark:text-white hover:text-primary transition-colors"
                      >
                        {video.external_channel_display_name}
                      </a>
                      <p class="text-taupe text-sm">
                        <a
                          href={`/channels/${video.channel_id}`}
                          class="no-underline text-taupe hover:text-primary transition-colors"
                        >
                          {video.channel_custom_id}
                        </a>
                        {" · "}
                        {formatSubscriberCount(video.external_channel_subscribers_count)}{" "}
                        {t("channelDetail.subscribers")}
                      </p>
                    </div>
                  </div>
              </div>
              <div class="flex items-center gap-6 mt-3 pb-3 border-b border-border-light dark:border-border-dark">
                  <button
                    class={`flex flex-col items-center gap-0.5 bg-transparent border-none transition-colors ${
                      markingWatchLater
                        ? "text-text-muted-light dark:text-text-muted-dark cursor-not-allowed opacity-50"
                        : markedWatchLater
                          ? "text-primary cursor-pointer hover:text-primary/80"
                          : "text-charcoal dark:text-white cursor-pointer hover:text-primary"
                    }`}
                    disabled={markingWatchLater}
                    onClick={() => requireAuth(async () => {
                      if (markingWatchLater || !videoId) return;
                      setMarkingWatchLater(true);
                      try {
                        if (markedWatchLater) {
                          await getPlaylist().deleteVideosVideoIdWatchLater(videoId);
                          setMarkedWatchLater(false);
                        } else {
                          await getPlaylist().postVideosVideoIdWatchLater(videoId);
                          setMarkedWatchLater(true);
                        }
                      } finally {
                        setMarkingWatchLater(false);
                      }
                    })}
                  >
                    <Icon name="schedule" class="text-lg" />
                    <span class="text-[10px] font-semibold">{t("videoPlayer.watchLater")}</span>
                  </button>
                  <button
                    class={`flex flex-col items-center gap-0.5 bg-transparent border-none transition-colors ${
                      markingWatched
                        ? "text-text-muted-light dark:text-text-muted-dark cursor-not-allowed opacity-50"
                        : markedWatched
                          ? "text-primary cursor-pointer hover:text-primary/80"
                          : "text-charcoal dark:text-white cursor-pointer hover:text-primary"
                    }`}
                    disabled={markingWatched}
                    onClick={() => requireAuth(async () => {
                      if (markingWatched || !videoId) return;
                      setMarkingWatched(true);
                      try {
                        if (markedWatched) {
                          await getHistory().deleteVideosVideoIdWatched(videoId);
                          setMarkedWatched(false);
                        } else {
                          await getHistory().postVideosVideoIdWatched(videoId);
                          setMarkedWatched(true);
                        }
                      } finally {
                        setMarkingWatched(false);
                      }
                    })}
                  >
                    <Icon name="check_circle" class="text-lg" />
                    <span class="text-[10px] font-semibold">{t("videoCard.markWatchedButton")}</span>
                  </button>
                  <button
                    class={`flex flex-col items-center gap-0.5 bg-transparent border-none transition-colors cursor-pointer ${
                      isLooping
                        ? "text-primary hover:text-primary/80"
                        : "text-charcoal dark:text-white hover:text-primary"
                    }`}
                    onClick={toggleLoop}
                  >
                    <Icon name="repeat" class="text-lg" />
                    <span class="text-[10px] font-semibold">{t("videoPlayer.loop")}</span>
                  </button>
                  <button
                    class="flex flex-col items-center gap-0.5 bg-transparent border-none text-charcoal dark:text-white hover:text-primary transition-colors cursor-pointer"
                    onClick={() => requireAuth(openPlaylistDialog)}
                  >
                    <Icon name="playlist_add" class="text-lg" />
                    <span class="text-[10px] font-semibold">{t("videoPlayer.playlist")}</span>
                  </button>
                  <a
                    href={`https://www.youtube.com/watch?v=${video.external_video_id}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="flex flex-col items-center gap-0.5 text-charcoal dark:text-white hover:text-primary transition-colors no-underline"
                  >
                    <Icon name="open_in_new" class="text-lg" />
                    <span class="text-[10px] font-semibold">{t("videoPlayer.openOnYouTube")}</span>
                  </a>
                </div>

              {/* Description */}
              {video.external_video_description && (
                <div class="mt-6">
                  <div class="bg-border-light/50 dark:bg-[#332e27]/30 p-6 rounded-xl">
                    <p class="text-base text-charcoal dark:text-white mb-3">
                      {new Date(video.external_video_created_at).toLocaleDateString("ja-JP", { year: "numeric", month: "2-digit", day: "2-digit" }).replaceAll("-", "/")}
                    </p>
                    <div
                      ref={descRef}
                      class={`text-charcoal dark:text-white/80 leading-relaxed whitespace-pre-line overflow-hidden ${isDescExpanded ? "" : "max-h-[4.875rem]"}`}
                    >
                        <Linkify text={video.external_video_description} onTimestamp={seekTo} />
                    </div>
                    {(descOverflows || isDescExpanded) && (
                      <button
                        class="mt-3 text-sm font-semibold text-primary hover:text-primary/80 transition-colors bg-transparent border-none cursor-pointer p-0"
                        onClick={() => setIsDescExpanded((v) => !v)}
                      >
                        {isDescExpanded
                          ? t("channelDetail.showLess")
                          : t("channelDetail.showMore")}
                      </button>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Sidebar */}
          {playlistId && <aside class="w-full xl:w-[420px] shrink-0 flex flex-col gap-8 px-4 sm:px-0">
            {/* Playlist sidebar */}
            {playlistId && playlistLoading && (
              <div class="bg-card-light dark:bg-card-dark rounded-2xl border border-border-light dark:border-border-dark p-8">
                <LoadingSpinner size="sm" />
              </div>
            )}

            {playlistId && !playlistLoading && playlistVideos.length > 0 && (
              <div class="bg-card-light dark:bg-card-dark rounded-2xl border border-border-light dark:border-border-dark flex flex-col overflow-hidden">
                {/* Playlist header */}
                <div class="p-4 border-b border-border-light dark:border-border-dark flex items-center justify-between gap-3">
                  <a href={`/playlists/${playlistId}`} class="flex items-center gap-2 min-w-0 no-underline group/pl-title">
                    <Icon name="playlist_play" class="text-primary text-xl flex-shrink-0" />
                    <div class="min-w-0">
                      <h2 class="font-bold text-sm tracking-tight truncate text-charcoal dark:text-white group-hover/pl-title:text-primary transition-colors">
                        {playlistInfo?.playlist_title ?? t("videoPlayer.curatedPlaylist")}
                      </h2>
                      <p class="text-[11px] text-text-muted-light dark:text-text-muted-dark">
                        {(playlistVideos.findIndex((v) => v.video_id === videoId) + 1) || "—"}{" "}
                        / {playlistInfo?.playlist_video_count ?? playlistVideos.length}
                      </p>
                    </div>
                  </a>
                  {playlistInfo?.playlist_type === "normal" && (
                    <div class="flex items-center gap-1 flex-shrink-0">
                      <button
                        class="p-1.5 rounded-lg text-text-muted-light dark:text-text-muted-dark hover:text-primary hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer bg-transparent border-none"
                        title={t("playlistDetail.addVideo")}
                        onClick={() => setShowImportVideo(true)}
                      >
                        <Icon name="add" class="text-[20px]" />
                      </button>
                      <button
                        class="p-1.5 rounded-lg text-text-muted-light dark:text-text-muted-dark hover:text-primary hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer bg-transparent border-none"
                        title={t("playlistDetail.edit")}
                        onClick={() => setShowEditPlaylist(true)}
                      >
                        <Icon name="edit" class="text-[20px]" />
                      </button>
                    </div>
                  )}
                </div>

                {/* Video list */}
                <div
                  class="overflow-y-auto max-h-[480px]"
                  onScroll={(e) => {
                    const el = e.currentTarget;
                    if (el.scrollHeight - el.scrollTop - el.clientHeight < 200) {
                      loadMorePlaylistVideos();
                    }
                  }}
                >
                  {playlistVideos.map((pv, idx) => {
                    const isCurrent = pv.video_id === videoId;
                    return (
                      <div key={pv.video_id} class="group/pv relative">
                        <a
                          href={buildWatchUrl(pv.video_id, pv.last_watch_seconds, playlistId!)}
                          class={`flex gap-3 p-3 pr-9 no-underline transition-colors ${
                            isCurrent
                              ? "bg-primary/10 dark:bg-primary/20"
                              : "hover:bg-black/5 dark:hover:bg-white/5"
                          }`}
                        >
                          <span class="text-xs text-text-muted-light dark:text-text-muted-dark w-5 flex-shrink-0 flex items-center justify-center pt-1">
                            {isCurrent ? (
                              <Icon name="play_arrow" class="text-primary text-base" />
                            ) : (
                              idx + 1
                            )}
                          </span>
                          <div class="relative w-20 aspect-video flex-shrink-0 rounded-md overflow-hidden bg-gray-200 dark:bg-gray-800">
                            <img
                              src={pv.external_video_thumbnail_url}
                              alt=""
                              loading="lazy"
                              class="absolute inset-0 w-full h-full object-cover"
                            />
                            <span class="absolute bottom-0.5 right-0.5 bg-black/80 text-white text-[10px] font-bold px-1 py-0.5 rounded">
                              {formatDuration(pv.external_video_length_seconds)}
                            </span>
                            {pv.last_watch_seconds != null && pv.external_video_length_seconds > 0 && (
                              <div class="absolute bottom-0 left-0 right-0 h-1 bg-white/30">
                                <div
                                  class="h-full bg-primary"
                                  style={{ width: `${Math.min(100, (pv.last_watch_seconds / pv.external_video_length_seconds) * 100)}%` }}
                                />
                              </div>
                            )}
                          </div>
                          <div class="flex flex-col justify-center min-w-0 flex-1">
                            <p
                              class={`text-xs font-semibold leading-tight line-clamp-2 ${
                                isCurrent
                                  ? "text-primary"
                                  : "text-charcoal dark:text-white"
                              }`}
                            >
                              {pv.external_video_title}
                            </p>
                            <p class="text-[10px] text-text-muted-light dark:text-text-muted-dark mt-0.5 truncate">
                              {pv.external_channel_display_name}
                            </p>
                          </div>
                        </a>
                        {playlistInfo?.playlist_type === "normal" && (
                          <button
                            class="absolute right-1.5 top-1/2 -translate-y-1/2 p-1 rounded-md text-text-muted-light dark:text-text-muted-dark hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors cursor-pointer bg-transparent border-none hidden group-hover/pv:block focus:block"
                            title={t("playlistDetail.removeVideo")}
                            disabled={removingVideoId === pv.video_id}
                            onClick={(e) => {
                              e.preventDefault();
                              handleRemoveFromPlaylist(pv.video_id);
                            }}
                          >
                            <Icon name="close" class="text-[16px]" />
                          </button>
                        )}
                      </div>
                    );
                  })}
                  {playlistLoadingMore && <LoadingSpinner size="sm" className="py-4" />}
                </div>
              </div>
            )}

          </aside>}
        </div>
      </div>
      {/* Add to Playlist Dialog */}
      <Dialog
        open={showPlaylistDialog}
        onClose={() => setShowPlaylistDialog(false)}
        ariaLabel={t("videoPlayer.addToPlaylist")}
        showCloseButton
        panelClass="max-h-[80vh] flex flex-col p-6"
      >
            <h2 class="text-xl font-bold text-charcoal dark:text-white mb-4">
              {t("videoPlayer.addToPlaylist")}
            </h2>
            <div class="flex flex-col gap-2 overflow-y-auto flex-1">
              {playlistDialogLoading ? (
                <LoadingSpinner size="sm" className="py-8" />
              ) : userPlaylists.length === 0 ? (
                <p class="text-sm text-text-muted-light dark:text-text-muted-dark text-center py-8">
                  {t("videoPlayer.noPlaylists")}
                </p>
              ) : (
                userPlaylists.map((pl) => {
                  const alreadyAdded = addedPlaylistIds.has(pl.playlist_id);
                  return (
                    <button
                      key={pl.playlist_id}
                      class={`flex items-center gap-3 p-3 rounded-xl border transition-all w-full text-left ${
                        alreadyAdded || addedToPlaylist === pl.playlist_id
                          ? "bg-green-50 dark:bg-green-900/20 border-green-300 dark:border-green-800 cursor-default opacity-70"
                          : failedToAddPlaylist === pl.playlist_id
                            ? "bg-red-50 dark:bg-red-900/20 border-red-300 dark:border-red-800 cursor-pointer"
                            : "bg-background-light dark:bg-neutral-800 border-border-light dark:border-border-dark hover:border-primary/30 cursor-pointer"
                      }`}
                      disabled={alreadyAdded || addingToPlaylist === pl.playlist_id}
                      onClick={() => handleAddToPlaylist(pl.playlist_id)}
                    >
                      <Icon name="playlist_play" class="text-primary text-xl" />
                      <div class="min-w-0 flex-1">
                        <p class="text-sm font-semibold text-charcoal dark:text-white truncate">
                          {pl.playlist_title}
                        </p>
                        <p class="text-[11px] text-text-muted-light dark:text-text-muted-dark">
                          {t("playlists.videoCount", {
                            count: pl.playlist_video_count,
                          })}
                        </p>
                      </div>
                      {alreadyAdded || addedToPlaylist === pl.playlist_id ? (
                        <Icon name="check_circle" class="text-green-600 dark:text-green-400 text-xl" />
                      ) : failedToAddPlaylist === pl.playlist_id ? (
                        <Icon name="error" class="text-red-500 text-xl" />
                      ) : addingToPlaylist === pl.playlist_id ? (
                        <Icon name="progress_activity" class="animate-spin text-primary text-xl" />
                      ) : (
                        <Icon name="add" class="text-text-muted-light dark:text-text-muted-dark text-xl" />
                      )}
                    </button>
                  );
                })
              )}
            </div>
      </Dialog>

      {playlistId && (
        <AddVideoDialog
          open={showImportVideo}
          playlistId={playlistId}
          onClose={() => setShowImportVideo(false)}
          onAdded={reloadPlaylist}
        />
      )}

      {playlistId && playlistInfo && (
        <EditPlaylistDialog
          open={showEditPlaylist}
          playlist={playlistInfo}
          onClose={() => setShowEditPlaylist(false)}
          onSaved={(updated) =>
            setPlaylistInfo((prev) => (prev ? { ...prev, ...updated } : prev))
          }
        />
      )}
      <AuthPromptDialog open={showAuthPrompt} onClose={closeAuthPrompt} />
    </DashboardLayout>
  );
}

export default function VideoPlayer() {
  return <VideoPlayerContent />;
}
