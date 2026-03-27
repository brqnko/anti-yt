import { useState, useEffect, useRef, useCallback } from "preact/hooks";
import type { RefObject } from "preact";
import { getHistory } from "../../api/generated/history";
import { getCookie } from "../../utils/cookie";
import { getCachedVisitorId } from "../../api/axios-instance";

const HEARTBEAT_INTERVAL_MS = 60_000;
const HEARTBEAT_DEBOUNCE_MS = 2_000;

/** Fire-and-forget heartbeat that survives page unload / navigation. */
function sendBeaconHeartbeat(videoId: string, positionSeconds: number): void {
  const url = `/api/v1/videos/${encodeURIComponent(videoId)}/heartbeats`;
  const body = JSON.stringify({ current_position_seconds: positionSeconds });
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const csrfToken = getCookie("csrf_token");
  if (csrfToken) headers["x-csrf-token"] = csrfToken;
  const visitorId = getCachedVisitorId();
  if (visitorId) headers["X-Device-Fingerprint"] = visitorId;

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
  playerState: number;
  currentTimeRef: RefObject<number>;
  togglePlay: () => void;
}

/**
 * Manages watch-time heartbeats for the video player.
 *
 * Sends periodic heartbeats while playing, handles final heartbeat on
 * pause/unload, and counts down remaining daily screen time.
 */
export function useHeartbeat({
  videoId,
  playerState,
  currentTimeRef,
  togglePlay,
}: UseHeartbeatOptions): {
  remainingSeconds: number | null;
  tickRemaining: () => void;
} {
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null);

  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const heartbeatDebounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const heartbeatFailCountRef = useRef(0);
  const finalHeartbeatSentRef = useRef(false);
  const lastFinalSentAtRef = useRef(0);

  // Keep videoId accessible in callbacks without adding it to effect deps where not needed
  const videoIdRef = useRef(videoId);
  videoIdRef.current = videoId;

  // PlayerState.PLAYING === 1
  const isPlaying = playerState === 1;

  // Send a final heartbeat via keepalive fetch. Guarded to fire at most once per pause/unload.
  const sendFinalHeartbeat = useCallback(() => {
    if (finalHeartbeatSentRef.current) return;
    const now = Date.now();
    if (now - lastFinalSentAtRef.current < 10_000) return;
    const vid = videoIdRef.current;
    if (vid) {
      finalHeartbeatSentRef.current = true;
      lastFinalSentAtRef.current = now;
      sendBeaconHeartbeat(vid, Math.floor(currentTimeRef.current));
    }
  }, [currentTimeRef]);

  // Send final heartbeat on tab close / hard navigation while playing
  useEffect(() => {
    if (!videoId || !isPlaying) return;

    const handler = () => sendFinalHeartbeat();
    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, [videoId, isPlaying, sendFinalHeartbeat]);

  // Heartbeat for watch time tracking (with debounce on play start)
  useEffect(() => {
    if (!videoId || !isPlaying) {
      if (heartbeatDebounceRef.current) {
        clearTimeout(heartbeatDebounceRef.current);
        heartbeatDebounceRef.current = null;
      }
      if (heartbeatRef.current) {
        clearInterval(heartbeatRef.current);
        heartbeatRef.current = null;
      }
      return;
    }

    // Reset guard — playback resumed, allow a new final heartbeat on next pause/unload
    finalHeartbeatSentRef.current = false;

    const sendHeartbeat = () => {
      getHistory()
        .postVideosVideoIdHeartbeats(videoId, {
          current_position_seconds: Math.floor(currentTimeRef.current),
        })
        .then((res) => {
          heartbeatFailCountRef.current = 0;
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
          const status = (err as { response?: { status?: number } })?.response?.status;
          console.warn(`[heartbeat] request failed (status=${status ?? "unknown"})`);
          heartbeatFailCountRef.current += 1;
          if (heartbeatFailCountRef.current >= 3) {
            console.warn("[heartbeat] 3 consecutive failures — pausing playback");
            togglePlay();
          }
        });
    };

    // Debounce: wait before sending first heartbeat to avoid bursts from rapid pause/play
    heartbeatDebounceRef.current = setTimeout(() => {
      sendHeartbeat();
      heartbeatRef.current = setInterval(sendHeartbeat, HEARTBEAT_INTERVAL_MS);
    }, HEARTBEAT_DEBOUNCE_MS);

    return () => {
      if (heartbeatDebounceRef.current) {
        clearTimeout(heartbeatDebounceRef.current);
        heartbeatDebounceRef.current = null;
      }
      if (heartbeatRef.current) {
        clearInterval(heartbeatRef.current);
        heartbeatRef.current = null;
      }
      sendFinalHeartbeat();
    };
  }, [videoId, isPlaying, togglePlay, sendFinalHeartbeat, currentTimeRef]);

  // Called by the player's syncTick so remaining countdown shares the same timer.
  const tickRemaining = useCallback(() => {
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
  }, [togglePlay]);

  return { remainingSeconds, tickRemaining };
}
