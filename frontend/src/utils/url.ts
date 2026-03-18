export function buildWatchUrl(
  videoId: string,
  watchedSeconds?: number,
): string {
  const base = `/watch/${videoId}`;
  return watchedSeconds ? `${base}?t=${Math.floor(watchedSeconds)}` : base;
}
