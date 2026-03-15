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
      window.onYouTubeIframeAPIReady = () => {
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
}: UseYouTubePlayerOptions) {
  const playerRef = useRef<YT.Player | null>(null);
  const [isReady, setIsReady] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const [playerState, setPlayerState] = useState<PlayerStateValue>(PlayerState.UNSTARTED);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolumeState] = useState(() => loadPreference(STORAGE_KEY_VOLUME, 100));
  const [isMuted, setIsMuted] = useState(() => loadPreference(STORAGE_KEY_MUTED, false));
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Sync time periodically while playing
  const startTimeSync = useCallback(() => {
    if (timerRef.current) clearInterval(timerRef.current);
    timerRef.current = setInterval(() => {
      const p = playerRef.current;
      if (p?.getCurrentTime) {
        setCurrentTime(p.getCurrentTime());
      }
    }, 250);
  }, []);

  const stopTimeSync = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  useEffect(() => {
    let cancelled = false;

    setLoadError(false);
    loadIframeAPI().then(() => {
      if (cancelled) return;

      const player = new window.YT.Player(containerId, {
        videoId,
        playerVars: {
          autoplay: 0,
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

            onReady?.();
          },
          onStateChange: (e: YT.OnStateChangeEvent) => {
            if (cancelled) return;
            const state = e.data as PlayerStateValue;
            setPlayerState(state);
            onStateChange?.(state);

            if (state === PlayerState.PLAYING) {
              setDuration(player.getDuration());
              startTimeSync();
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
      setIsReady(false);
    };
  }, [videoId, containerId]);

  const play = useCallback(() => playerRef.current?.playVideo(), []);
  const pause = useCallback(() => playerRef.current?.pauseVideo(), []);
  const togglePlay = useCallback(() => {
    const p = playerRef.current;
    if (!p) return;
    const state = p.getPlayerState();
    if (state === PlayerState.PLAYING) {
      p.pauseVideo();
    } else {
      p.playVideo();
    }
  }, []);

  const seekTo = useCallback((seconds: number) => {
    playerRef.current?.seekTo(seconds, true);
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
    duration,
    volume,
    isMuted,
    play,
    pause,
    togglePlay,
    seekTo,
    setVolume,
    toggleMute,
  };
}
