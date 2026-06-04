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
