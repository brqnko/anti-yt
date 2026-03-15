export function formatDuration(totalSeconds: number): string {
  const h = Math.floor(totalSeconds / 3600);
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = Math.floor(totalSeconds % 60);
  const mm = String(m).padStart(2, "0");
  const ss = String(s).padStart(2, "0");
  return h > 0 ? `${h}:${mm}:${ss}` : `${mm}:${ss}`;
}

export function formatTimeAgo(
  dateStr: string,
  t: (key: string, opts?: object) => string,
): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);
  const weeks = Math.floor(days / 7);
  const months = Math.floor(days / 30);
  const years = Math.floor(days / 365);

  if (years > 0) return t("dashboard.timeAgo.years", { count: years });
  if (months > 0) return t("dashboard.timeAgo.months", { count: months });
  if (weeks > 0) return t("dashboard.timeAgo.weeks", { count: weeks });
  if (days > 0) return t("dashboard.timeAgo.days", { count: days });
  if (hours > 0) return t("dashboard.timeAgo.hours", { count: hours });
  if (minutes > 0) return t("dashboard.timeAgo.minutes", { count: minutes });
  return t("dashboard.timeAgo.justNow");
}

export function formatSubscriberCount(count: number): string {
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K`;
  return String(count);
}
