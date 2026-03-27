import { useEffect, useRef, useCallback, useState } from "preact/hooks";

declare global {
  interface Window {
    YT: typeof YT;
    onYouTubeIframeAPIReady: (() => void) | undefined;
  }
}

export const PlayerState = {
  UNSTARTED: -1,
  ENDED: 0,
  PLAYING: 1,
  PAUSED: 2,
  BUFFERING: 3,
  CUED: 5,
} as const;

type PlayerStateValue = (typeof PlayerState)[keyof typeof PlayerState];

interface UseYouTubePlayerOptions {
  videoId: string;
  containerId: string;
  onStateChange?: (state: PlayerStateValue) => void;
  onReady?: () => void;
  /** Called on every low-frequency sync tick (1 s interval) while playing. */
  onSyncTick?: () => void;
}

const IFRAME_API_TIMEOUT_MS = 10_000;

// localStorage keys for persisting player preferences
const STORAGE_KEY_VOLUME = "yt-player-volume";
const STORAGE_KEY_MUTED = "yt-player-muted";

function loadPreference<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    if (raw === null) return fallback;
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

function savePreference(key: string, value: unknown): void {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // storage full or unavailable — silently ignore
  }
}

function loadIframeAPI(): Promise<void> {
  return new Promise((resolve, reject) => {
    if (window.YT?.Player) {
      resolve();
      return;
    }

    const existing = document.getElementById("yt-iframe-api");
    if (existing) {
      const prev = window.onYouTubeIframeAPIReady;
      const fallbackTimeout = setTimeout(() => {
        // Script tag exists but API never loaded — remove and reject
        existing.remove();
        reject(new Error("YouTube IFrame API load timed out (existing script)"));
      }, IFRAME_API_TIMEOUT_MS);
      window.onYouTubeIframeAPIReady = () => {
        clearTimeout(fallbackTimeout);
        prev?.();
        resolve();
      };
      return;
    }

    const timeout = setTimeout(() => {
      reject(new Error("YouTube IFrame API load timed out"));
    }, IFRAME_API_TIMEOUT_MS);

    window.onYouTubeIframeAPIReady = () => {
      clearTimeout(timeout);
      resolve();
    };
    const script = document.createElement("script");
    script.id = "yt-iframe-api";
    script.src = "https://www.youtube.com/iframe_api";
    script.onerror = () => {
      clearTimeout(timeout);
      document.getElementById("yt-iframe-api")?.remove();
      reject(new Error("Failed to load YouTube IFrame API"));
    };
    document.head.appendChild(script);
  });
}

export function useYouTubePlayer({
  videoId,
  containerId,
  onStateChange,
  onReady,
  onSyncTick,
}: UseYouTubePlayerOptions) {
  const playerRef = useRef<YT.Player | null>(null);
  const [isReady, setIsReady] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const [playerState, setPlayerState] = useState<PlayerStateValue>(PlayerState.UNSTARTED);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolumeState] = useState(() => loadPreference(STORAGE_KEY_VOLUME, 100));
  const [isMuted, setIsMuted] = useState(() => loadPreference(STORAGE_KEY_MUTED, false));
  const rafRef = useRef<number | null>(null);
  const currentTimeRef = useRef(0);
  const lastRenderedSecondRef = useRef(-1);
  const loadedVideoIdRef = useRef("");
  const videoIdRef = useRef(videoId);
  videoIdRef.current = videoId;

  // Keep callback refs stable to avoid re-creating the player on every render
  const onReadyRef = useRef(onReady);
  onReadyRef.current = onReady;
  const onStateChangeRef = useRef(onStateChange);
  onStateChangeRef.current = onStateChange;
  const onSyncTickRef = useRef(onSyncTick);
  onSyncTickRef.current = onSyncTick;

  // Sync time while playing.
  // Two modes: rAF (high-frequency, for visible progress bar) and interval (low-frequency fallback).
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const highFreqRef = useRef(false);
  const lastOnSyncTickAtRef = useRef(0);

  const syncTick = useCallback(() => {
    const p = playerRef.current;
    if (p?.getCurrentTime) {
      const t = p.getCurrentTime();
      currentTimeRef.current = t;
      const sec5 = Math.floor(t / 5);
      if (sec5 !== lastRenderedSecondRef.current) {
        lastRenderedSecondRef.current = sec5;
        setCurrentTime(t);
      }
    }
    // Fire onSyncTick at most once per second (works in both rAF and interval modes).
    const now = performance.now();
    if (now - lastOnSyncTickAtRef.current >= 1000) {
      lastOnSyncTickAtRef.current = now;
      onSyncTickRef.current?.();
    }
  }, []);

  const stopTimeSync = useCallback(() => {
    if (rafRef.current !== null) {
      cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
    }
    if (intervalRef.current !== null) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    highFreqRef.current = false;
  }, []);

  const startTimeSync = useCallback((highFreq: boolean) => {
    stopTimeSync();
    if (highFreq) {
      highFreqRef.current = true;
      const tick = () => {
        syncTick();
        rafRef.current = requestAnimationFrame(tick);
      };
      rafRef.current = requestAnimationFrame(tick);
    } else {
      highFreqRef.current = false;
      syncTick();
      intervalRef.current = setInterval(syncTick, 1000);
    }
  }, [syncTick, stopTimeSync]);

  // Allow the parent to switch between high-freq (rAF) and low-freq (interval) sync
  const setHighFreqSync = useCallback((enabled: boolean) => {
    // Only switch if currently syncing and mode actually changed
    if (rafRef.current === null && intervalRef.current === null) return;
    if (enabled === highFreqRef.current) return;
    startTimeSync(enabled);
  }, [startTimeSync]);

  // Create the player once per container. Video changes are handled by a
  // separate effect that calls loadVideoById so the player instance (and its
  // autoplay permission) survives across playlist items – this avoids the
  // browser blocking autoplay when the tab is in the background.
  useEffect(() => {
    let cancelled = false;

    setLoadError(false);

    // Wait for the container element to exist in the DOM.
    // The VideoPlayer component conditionally renders the container div
    // (hidden behind a loading spinner until video data loads), so the
    // element may not be present when this effect first runs.
    function waitForContainer(): Promise<void> {
      return new Promise((resolve, reject) => {
        if (document.getElementById(containerId)) {
          resolve();
          return;
        }
        let elapsed = 0;
        const interval = 100;
        const maxWait = 15_000;
        const timer = setInterval(() => {
          if (cancelled) {
            clearInterval(timer);
            reject(new Error("cancelled"));
            return;
          }
          elapsed += interval;
          if (document.getElementById(containerId)) {
            clearInterval(timer);
            resolve();
          } else if (elapsed >= maxWait) {
            clearInterval(timer);
            reject(new Error("Container element not found"));
          }
        }, interval);
      });
    }

    Promise.all([loadIframeAPI(), waitForContainer()]).then(() => {
      if (cancelled) return;

      // Re-read the latest videoId — the ref may have been updated while
      // we were waiting for the iframe API and/or container element.
      const currentVideoId = videoIdRef.current;
      const player = new window.YT.Player(containerId, {
        videoId: currentVideoId || undefined,
        playerVars: {
          autoplay: 1,
          modestbranding: 1,
          rel: 0,
          controls: 0,
          disablekb: 1,
          iv_load_policy: 3,
          playsinline: 1,
        },
        events: {
          onReady: () => {
            if (cancelled) return;
            setIsReady(true);
            setDuration(player.getDuration());
            loadedVideoIdRef.current = currentVideoId;

            // Restore saved volume & mute preferences
            const savedVolume = loadPreference(STORAGE_KEY_VOLUME, 100);
            const savedMuted = loadPreference(STORAGE_KEY_MUTED, false);
            player.setVolume(savedVolume);
            setVolumeState(savedVolume);
            if (savedMuted) {
              player.mute();
            } else {
              player.unMute();
            }
            setIsMuted(savedMuted);

            onReadyRef.current?.();
          },
          onStateChange: (e: YT.OnStateChangeEvent) => {
            if (cancelled) return;
            const state = e.data as PlayerStateValue;
            setPlayerState(state);
            onStateChangeRef.current?.(state);

            if (state === PlayerState.PLAYING) {
              setDuration(player.getDuration());
              startTimeSync(false);
            } else {
              stopTimeSync();
              if (player.getCurrentTime) {
                setCurrentTime(player.getCurrentTime());
              }
            }
          },
        },
      });

      playerRef.current = player;
    }).catch(() => {
      if (!cancelled) setLoadError(true);
    });

    return () => {
      cancelled = true;
      stopTimeSync();
      playerRef.current?.destroy();
      playerRef.current = null;
      loadedVideoIdRef.current = "";
      setIsReady(false);
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps -- startTimeSync/stopTimeSync are stable (useCallback with []), callbacks via refs
  }, [containerId]);

  // Load a new video into the existing player when videoId changes.
  useEffect(() => {
    if (!videoId || videoId === loadedVideoIdRef.current) return;
    const p = playerRef.current;
    if (p && isReady) {
      p.loadVideoById(videoId);
      loadedVideoIdRef.current = videoId;
      setCurrentTime(0);
      setDuration(0);
    }
  }, [videoId, isReady]);

  const play = useCallback(() => playerRef.current?.playVideo(), []);
  const pause = useCallback(() => playerRef.current?.pauseVideo(), []);
  const togglePlay = useCallback(() => {
    const p = playerRef.current;
    if (!p || typeof p.getPlayerState !== "function") return;
    const state = p.getPlayerState();
    if (state === PlayerState.PLAYING) {
      p.pauseVideo();
    } else {
      p.playVideo();
    }
  }, []);

  const seekTo = useCallback((seconds: number) => {
    playerRef.current?.seekTo(seconds, true);
    currentTimeRef.current = seconds;
    lastRenderedSecondRef.current = Math.floor(seconds);
    setCurrentTime(seconds);
  }, []);

  const setVolume = useCallback((val: number) => {
    playerRef.current?.setVolume(val);
    setVolumeState(val);
    savePreference(STORAGE_KEY_VOLUME, val);
    if (val > 0 && playerRef.current?.isMuted()) {
      playerRef.current.unMute();
      setIsMuted(false);
      savePreference(STORAGE_KEY_MUTED, false);
    }
  }, []);

  const toggleMute = useCallback(() => {
    const p = playerRef.current;
    if (!p) return;
    if (p.isMuted()) {
      p.unMute();
      setIsMuted(false);
      savePreference(STORAGE_KEY_MUTED, false);
    } else {
      p.mute();
      setIsMuted(true);
      savePreference(STORAGE_KEY_MUTED, true);
    }
  }, []);

  return {
    isReady,
    loadError,
    playerState,
    currentTime,
    currentTimeRef,
    duration,
    volume,
    isMuted,
    play,
    pause,
    togglePlay,
    seekTo,
    setVolume,
    toggleMute,
    setHighFreqSync,
  };
}
