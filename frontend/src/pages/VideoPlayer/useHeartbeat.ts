import { useState, useEffect, useRef, useCallback } from "preact/hooks";
import type { RefObject } from "preact";
import { AxiosError } from "axios";
import { getHistory } from "../../api/generated/history";
import { getCookie } from "../../utils/cookie";

const HEARTBEAT_INTERVAL_TICKS = 60; // 60 ticks = 60 seconds (onSyncTick fires every 1s)
const HEARTBEAT_DEBOUNCE_TICKS = 2;

/** Fire-and-forget heartbeat that survives page unload / navigation. */
function sendBeaconHeartbeat(videoId: string, positionSeconds: number, playlistId: string | null): void {
  const url = `/api/v1/videos/${encodeURIComponent(videoId)}/heartbeats`;
  const payload: Record<string, unknown> = { current_position_seconds: positionSeconds };
  if (playlistId) payload.playlist_id = playlistId;
  const body = JSON.stringify(payload);
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const csrfToken = getCookie("csrf_token");
  if (csrfToken) headers["x-csrf-token"] = csrfToken;
  headers["X-Timezone"] = Intl.DateTimeFormat().resolvedOptions().timeZone;

  try {
    fetch(url, {
      method: "POST",
      headers,
      body,
      credentials: "include",
      keepalive: true,
    });
  } catch {
    // best-effort — ignore failures on teardown
  }
}

interface UseHeartbeatOptions {
  videoId: string | null;
  playlistId: string | null;
  playerState: number;
  currentTimeRef: RefObject<number>;
  togglePlay: () => void;
}

/**
 * Manages watch-time heartbeats for the video player.
 *
 * Instead of running its own timers, exposes a `tick()` that should be called
 * from the player's onSyncTick (every ~1 s while playing). Sends a heartbeat
 * every HEARTBEAT_INTERVAL_TICKS ticks and counts down remaining daily screen time.
 */
export function useHeartbeat({
  videoId,
  playlistId,
  playerState,
  currentTimeRef,
  togglePlay,
}: UseHeartbeatOptions): {
  remainingSeconds: number | null;
  tick: () => void;
} {
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null);

  const tickCountRef = useRef(0);
  const finalHeartbeatSentRef = useRef(false);
  const lastFinalSentAtRef = useRef(0);

  const videoIdRef = useRef(videoId);
  videoIdRef.current = videoId;
  const playlistIdRef = useRef(playlistId);
  playlistIdRef.current = playlistId;

  // PlayerState.PLAYING === 1
  const isPlaying = playerState === 1;

  const sendFinalHeartbeat = useCallback(() => {
    if (finalHeartbeatSentRef.current) return;
    const now = Date.now();
    if (now - lastFinalSentAtRef.current < 10_000) return;
    const vid = videoIdRef.current;
    if (vid) {
      finalHeartbeatSentRef.current = true;
      lastFinalSentAtRef.current = now;
      sendBeaconHeartbeat(vid, Math.floor(currentTimeRef.current ?? 0), playlistIdRef.current);
    }
  }, [currentTimeRef]);

  // Send final heartbeat on tab close / hard navigation while playing
  useEffect(() => {
    if (!videoId || !isPlaying) return;

    const handler = () => sendFinalHeartbeat();
    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, [videoId, isPlaying, sendFinalHeartbeat]);

  // Reset tick counter & final-heartbeat guard when playback resumes
  useEffect(() => {
    if (isPlaying) {
      tickCountRef.current = 0;
      finalHeartbeatSentRef.current = false;
    }
  }, [isPlaying]);

  const sendHeartbeat = useCallback(() => {
    const vid = videoIdRef.current;
    if (!vid) return;
    getHistory()
      .postVideosVideoIdHeartbeats(vid, {
        current_position_seconds: Math.floor(currentTimeRef.current ?? 0),
        ...(playlistIdRef.current ? { playlist_id: playlistIdRef.current } : {}),
      })
      .then((res) => {
        const remaining = res.daily_remaining_seconds ?? null;
        setRemainingSeconds(remaining);
        if (remaining !== null && remaining <= 0) {
          togglePlay();
          window.dispatchEvent(
            new CustomEvent("screen-time:blocked", { detail: { reason: "limit_exceeded" } }),
          );
        }
      })
      .catch((err: unknown) => {
        const status = err instanceof AxiosError ? err.response?.status : undefined;
        console.warn(`[heartbeat] request failed (status=${status ?? "unknown"}) — pausing playback`);
        togglePlay();
      });
  }, [currentTimeRef, togglePlay]);

  // Send a heartbeat via the API when the video ends (ENDED === 0)
  useEffect(() => {
    if (playerState === 0 && videoId) {
      sendHeartbeat();
    }
  }, [playerState, videoId, sendHeartbeat]);

  // Called from onSyncTick every ~1 s while playing.
  const tick = useCallback(() => {
    tickCountRef.current += 1;

    // Heartbeat: first after debounce, then every interval
    const count = tickCountRef.current;
    if (count >= HEARTBEAT_DEBOUNCE_TICKS && (count - HEARTBEAT_DEBOUNCE_TICKS) % HEARTBEAT_INTERVAL_TICKS === 0) {
      sendHeartbeat();
    }

    setRemainingSeconds((prev) => {
      if (prev === null) return null;
      if (prev <= 1) {
        togglePlay();
        window.dispatchEvent(
          new CustomEvent("screen-time:blocked", { detail: { reason: "limit_exceeded" } }),
        );
        return 0;
      }
      return prev - 1;
    });
  }, [sendHeartbeat, togglePlay]);

  return { remainingSeconds, tick };
}
