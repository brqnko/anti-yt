export function buildWatchUrl(
  videoId: string,
  watchedSeconds?: number,
  playlistId?: string,
): string {
  const base = `/watch/${videoId}`;
  const params = new URLSearchParams();
  if (watchedSeconds) params.set("t", String(Math.floor(watchedSeconds)));
  if (playlistId) params.set("playlist", playlistId);
  const qs = params.toString();
  return qs ? `${base}?${qs}` : base;
}

const VIDEO_ID_RE = /^[A-Za-z0-9_-]{11}$/;
const CHANNEL_ID_RE = /^UC[A-Za-z0-9_-]{22}$/;

function parseTimeParam(value: string | null): number | undefined {
  if (!value) return undefined;
  const direct = Number(value);
  if (Number.isFinite(direct) && direct > 0) return Math.floor(direct);
  const m = /^(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?$/.exec(value);
  if (!m) return undefined;
  const [, h, mi, s] = m;
  const total =
    (h ? parseInt(h, 10) * 3600 : 0) +
    (mi ? parseInt(mi, 10) * 60 : 0) +
    (s ? parseInt(s, 10) : 0);
  return total > 0 ? total : undefined;
}

export function rewriteYouTubeUrl(raw: string): string | null {
  let url: URL;
  try {
    url = new URL(raw);
  } catch {
    return null;
  }
  const host = url.hostname.replace(/^www\./, "").toLowerCase();
  const isYouTube =
    host === "youtube.com" ||
    host === "m.youtube.com" ||
    host === "music.youtube.com";
  const isShortLink = host === "youtu.be";
  if (!isYouTube && !isShortLink) return null;

  if (isShortLink) {
    const id = url.pathname.replace(/^\/+/, "").split("/")[0];
    if (!VIDEO_ID_RE.test(id)) return null;
    const t = parseTimeParam(url.searchParams.get("t"));
    return buildWatchUrl(id, t);
  }

  const segments = url.pathname.split("/").filter(Boolean);
  const first = segments[0];

  if (url.pathname === "/watch") {
    const id = url.searchParams.get("v");
    if (!id || !VIDEO_ID_RE.test(id)) return null;
    const t = parseTimeParam(url.searchParams.get("t"));
    return buildWatchUrl(id, t);
  }

  if (first === "embed" || first === "v" || first === "live") {
    const id = segments[1];
    if (!id || !VIDEO_ID_RE.test(id)) return null;
    const t = parseTimeParam(url.searchParams.get("t") ?? url.searchParams.get("start"));
    return buildWatchUrl(id, t);
  }

  if (first === "channel") {
    const id = segments[1];
    if (!id || !CHANNEL_ID_RE.test(id)) return null;
    return `/channels/${id}`;
  }

  return null;
}
