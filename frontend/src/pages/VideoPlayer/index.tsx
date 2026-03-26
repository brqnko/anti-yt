import { useState, useEffect, useRef, useCallback } from "preact/hooks";
import { useRoute, useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { getVideo } from "../../api/generated/video";
import { getPlaylist } from "../../api/generated/playlist";
import { formatDuration, formatSubscriberCount } from "../../utils/format";
import { buildWatchUrl } from "../../utils/url";
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

function VideoPlayerContent() {
  const { t } = useTranslation();
  const { params } = useRoute();
  const { route } = useLocation();
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
  const [noteText, setNoteText] = useState("");
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isDescExpanded, setIsDescExpanded] = useState(false);
  const [controlsVisible, setControlsVisible] = useState(false);
  const [isSeeking, setIsSeeking] = useState(false);

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
  const [removingVideoId, setRemovingVideoId] = useState<string | null>(null);
  const [userPlaylists, setUserPlaylists] = useState<GetPlaylists200ItemsItem[]>([]);
  const [playlistDialogLoading, setPlaylistDialogLoading] = useState(false);
  const [addingToPlaylist, setAddingToPlaylist] = useState<string | null>(null);
  const [addedToPlaylist, setAddedToPlaylist] = useState<string | null>(null);
  const [failedToAddPlaylist, setFailedToAddPlaylist] = useState<string | null>(null);
  const [addedPlaylistIds, setAddedPlaylistIds] = useState<Set<string>>(new Set());
  const [showPlaylistDialog, setShowPlaylistDialog] = useState(false);

  // Refs for values used in keyboard handler / cleanup to avoid stale closures
  const durationRef = useRef(0);
  const volumeRef = useRef(100);

  // Refs for auto-play next video in playlist
  const playlistVideosRef = useRef(playlistVideos);
  playlistVideosRef.current = playlistVideos;
  const playlistIdRef = useRef(playlistId);
  playlistIdRef.current = playlistId;

  useTitle(video?.external_video_title ?? t("videoPlayer.pageTitle"));

  useEffect(() => {
    if (!videoId) return;
    setIsLoading(true);
    setError(false);
    getVideo()
      .getVideosVideoId(videoId)
      .then((res) => setVideo(res))
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
        const lastItem = videosRes.value.items[videosRes.value.items.length - 1];
        playlistCursorRef.current = lastItem?.video_id;
      }
    }).finally(() => setPlaylistLoading(false));
  }, [playlistId]);

  const loadMorePlaylistVideos = useCallback(async () => {
    if (playlistLoadingMore || !playlistHasNext || !playlistId) return;
    setPlaylistLoadingMore(true);
    try {
      const res = await getPlaylist().getPlaylistsPlaylistIdVideos(playlistId, {
        limit: PAGE_SIZES.PLAYLIST_VIDEOS,
        cursor: playlistCursorRef.current,
      });
      setPlaylistVideos((prev) => [...prev, ...res.items]);
      setPlaylistHasNext(res.has_next);
      const lastItem = res.items[res.items.length - 1];
      playlistCursorRef.current = lastItem?.video_id;
    } catch {
      setPlaylistHasNext(false);
    } finally {
      setPlaylistLoadingMore(false);
    }
  }, [playlistId, playlistLoadingMore, playlistHasNext]);

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
    loadError,
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
    setHighFreqSync,
  } = useYouTubePlayer({
    videoId: video?.external_video_id ?? "",
    containerId: PLAYER_CONTAINER_ID,
    onStateChange: handlePlayerStateChange,
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

  const { remainingSeconds } = useHeartbeat({
    videoId: video?.video_id ?? null,
    playerState,
    currentTimeRef,
    togglePlay,
  });

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
    const handler = () => setIsFullscreen(!!document.fullscreenElement);
    document.addEventListener("fullscreenchange", handler);
    return () => document.removeEventListener("fullscreenchange", handler);
  }, []);

  const toggleFullscreen = useCallback(() => {
    const el = playerWrapperRef.current;
    if (!el) return;
    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      el.requestFullscreen?.();
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

  if (isLoading) {
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
            href="/dashboard"
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
        <div class="max-w-[1536px] mx-auto px-6 py-8 flex flex-col xl:flex-row gap-8">
          {/* Main content */}
          <div class="flex-1 min-w-0">
            {/* YouTube Player */}
            <div
              ref={playerWrapperRef}
              class="w-full bg-black rounded-xl overflow-hidden ring-1 ring-white/10 relative aspect-video group/player"
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

              {/* Buffering spinner */}
              {isReady && playerState === PlayerState.BUFFERING && (
                <div class="absolute inset-0 z-20 flex items-center justify-center pointer-events-none">
                  <Icon name="progress_activity" class="text-5xl animate-spin text-white" />
                </div>
              )}

              {/* Loading overlay before player is ready */}
              {!isReady && (
                <div class="absolute inset-0 z-20">
                  <img
                    src={video.external_video_thumbnail_url}
                    alt=""
                    class="absolute inset-0 w-full h-full object-cover"
                  />
                  <div class="absolute inset-0 bg-black/30 flex items-center justify-center">
                    {loadError ? (
                      <Icon name="error" class="text-5xl text-white" />
                    ) : (
                      <Icon name="progress_activity" class="text-5xl animate-spin text-white" />
                    )}
                  </div>
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
            <div class="mt-8">
              <h1 class="text-xl font-bold leading-tight tracking-tight">
                {video.external_video_title}
              </h1>
              <div class="flex flex-col md:flex-row md:items-start justify-between gap-6 mt-4 pb-6 border-b border-border-light dark:border-border-dark">
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
                        {video.channel_custom_id}
                        {" · "}
                        {formatSubscriberCount(video.external_channel_subscribers_count)}{" "}
                        {t("channelDetail.subscribers")}
                      </p>
                    </div>
                  </div>
                <button
                  class="flex items-center gap-2 h-10 px-4 rounded-lg bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark hover:border-primary/30 text-charcoal dark:text-white font-semibold text-sm transition-all cursor-pointer hover:-translate-y-px flex-shrink-0 self-end"
                  onClick={openPlaylistDialog}
                >
                  {t("videoPlayer.addToPlaylist")}
                </button>
              </div>

              {/* Description */}
              {video.external_video_description && (
                <div class="mt-6">
                  <div class="bg-border-light/50 dark:bg-[#332e27]/30 p-6 rounded-xl">
                    <div
                      ref={descRef}
                      class={`text-charcoal dark:text-white/80 leading-relaxed whitespace-pre-line overflow-hidden ${isDescExpanded ? "" : "max-h-[4.875rem]"}`}
                    >
                        <Linkify text={video.external_video_description} />
                    </div>
                    {descOverflows && (
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
          <aside class="w-full xl:w-[420px] shrink-0 space-y-8">
            {/* Quick Notes */}
            <div class="bg-card-light dark:bg-card-dark rounded-2xl border border-border-light dark:border-border-dark flex flex-col overflow-hidden">
              <div class="p-4 border-b border-border-light dark:border-border-dark flex items-center justify-between">
                <h2 class="font-bold text-lg tracking-tight flex items-center gap-2">
                  <Icon name="edit_note" class="text-primary" />
                  {t("videoPlayer.quickNotes")}
                </h2>
              </div>
              <div class="px-4 py-2 flex items-center gap-1 border-b border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark/50">
                <button
                  class="p-1.5 rounded transition-colors bg-transparent border-none text-charcoal/30 dark:text-white/30 cursor-not-allowed"
                  title="Bold"
                  disabled
                >
                  <Icon name="format_bold" class="text-xl" />
                </button>
                <button
                  class="p-1.5 rounded transition-colors bg-transparent border-none text-charcoal/30 dark:text-white/30 cursor-not-allowed"
                  title="Italic"
                  disabled
                >
                  <Icon name="format_italic" class="text-xl" />
                </button>
                <button
                  class="p-1.5 rounded transition-colors bg-transparent border-none text-charcoal/30 dark:text-white/30 cursor-not-allowed"
                  title="Bullet List"
                  disabled
                >
                  <Icon name="format_list_bulleted" class="text-xl" />
                </button>
                <div class="w-px h-6 bg-border-light dark:bg-border-dark mx-1" />
                <button
                  class="p-1.5 rounded transition-colors bg-transparent border-none text-charcoal/30 dark:text-white/30 cursor-not-allowed"
                  title="Timestamp"
                  disabled
                >
                  <Icon name="schedule" class="text-xl" />
                </button>
              </div>
              <div class="relative">
                <textarea
                  class="w-full h-48 p-4 bg-transparent border-none focus:ring-0 focus:outline-none text-sm leading-relaxed resize-none text-charcoal dark:text-white"
                  placeholder={t("videoPlayer.notesPlaceholder")}
                  value={noteText}
                  onInput={(e) => setNoteText((e.target as HTMLTextAreaElement).value)}
                />
              </div>
              <div class="p-4 bg-background-light dark:bg-background-dark/50 border-t border-border-light dark:border-border-dark flex items-center justify-end">
                <button class="bg-primary text-white px-5 py-2 rounded-xl font-bold text-sm tracking-wide hover:opacity-90 transition-opacity border-none cursor-pointer">
                  {t("videoPlayer.saveNote")}
                </button>
              </div>
            </div>

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
                  <div class="flex items-center gap-2 min-w-0">
                    <Icon name="playlist_play" class="text-primary text-xl flex-shrink-0" />
                    <div class="min-w-0">
                      <h2 class="font-bold text-sm tracking-tight truncate">
                        {playlistInfo?.playlist_title ?? t("videoPlayer.curatedPlaylist")}
                      </h2>
                      <p class="text-[11px] text-text-muted-light dark:text-text-muted-dark">
                        {(playlistVideos.findIndex((v) => v.video_id === videoId) + 1) || "—"}{" "}
                        / {playlistInfo?.playlist_video_count ?? playlistVideos.length}
                      </p>
                    </div>
                  </div>
                  <a
                    href={`/playlists/${playlistId}`}
                    class="text-xs text-primary hover:text-primary/80 transition-colors font-medium no-underline flex-shrink-0"
                  >
                    {t("videoPlayer.viewPlaylist")}
                  </a>
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
                      </div>
                    );
                  })}
                  {playlistLoadingMore && <LoadingSpinner size="sm" className="py-4" />}
                </div>
              </div>
            )}

          </aside>
        </div>
      </div>
      {/* Add to Playlist Dialog */}
      {showPlaylistDialog && (
        <div
          class="fixed inset-0 z-50 flex items-center justify-center p-4"
          role="dialog"
          aria-modal="true"
          aria-label={t("videoPlayer.addToPlaylist")}
        >
          <div
            class="absolute inset-0 bg-black/60"
            onClick={() => setShowPlaylistDialog(false)}
          />
          <div class="relative bg-white dark:bg-[#2a2721] rounded-2xl ring-1 ring-black/10 dark:ring-white/10 border border-gray-100 dark:border-neutral-800 p-6 max-w-md w-full max-h-[80vh] flex flex-col">
            <button
              class="absolute top-4 right-4 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors bg-transparent border-none cursor-pointer"
              onClick={() => setShowPlaylistDialog(false)}
            >
              <Icon name="close" />
            </button>
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
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}

export default function VideoPlayer() {
  return (
    <ProtectedRoute>
      <VideoPlayerContent />
    </ProtectedRoute>
  );
}
