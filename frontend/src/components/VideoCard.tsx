import { useTranslation } from "react-i18next";
import { formatDuration, formatTimeAgo } from "../utils/format";
import { buildWatchUrl } from "../utils/url";

export interface VideoCardProps {
  videoId: string;
  thumbnailUrl: string;
  title: string;
  lengthSeconds: number;
  channel?: {
    channelId: string;
    iconUrl: string;
    displayName: string;
  };
  dateStr?: string;
  watchedAt?: string;
  watchedSeconds?: number;
  layout?: "card" | "row";
  playlistId?: string;
}

function VideoThumbnail({
  watchUrl,
  thumbnailUrl,
  title,
  lengthSeconds,
  progressPercent,
  size,
}: {
  watchUrl: string;
  thumbnailUrl: string;
  title: string;
  lengthSeconds: number;
  progressPercent: number;
  size: "card" | "row";
}) {
  return (
    <a
      href={watchUrl}
      class={`group/thumb relative aspect-video overflow-hidden bg-gray-200 dark:bg-gray-800 block no-underline ${
        size === "card" ? "rounded-xl" : "w-60 flex-shrink-0 rounded-lg"
      }`}
    >
      <img
        src={thumbnailUrl}
        alt={title}
        class="absolute inset-0 w-full h-full object-cover transition-transform duration-500 group-hover/thumb:scale-105"
      />
      <div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover/thumb:opacity-100 transition-opacity duration-300" />
      <span class="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-bold px-1.5 py-0.5 rounded">
        {formatDuration(lengthSeconds)}
      </span>
      <div class="absolute inset-0 flex items-center justify-center opacity-0 group-hover/thumb:opacity-100 transition-opacity duration-300 pointer-events-none">
        <div
          class={`rounded-full bg-primary/90 flex items-center justify-center text-white shadow-lg transform scale-90 group-hover/thumb:scale-100 transition-transform ${
            size === "card" ? "size-12" : "size-10"
          }`}
        >
          <span
            class={`material-symbols-outlined ${
              size === "card" ? "text-[28px] ml-1" : "text-[22px] ml-0.5"
            }`}
          >
            play_arrow
          </span>
        </div>
      </div>
      {progressPercent > 0 && (
        <div class="absolute bottom-0 left-0 right-0 h-1 bg-gray-400/50">
          <div
            class="h-full bg-primary"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
      )}
    </a>
  );
}

export function VideoCard({
  videoId,
  thumbnailUrl,
  title,
  lengthSeconds,
  channel,
  dateStr,
  watchedAt,
  watchedSeconds,
  layout = "card",
  playlistId,
}: VideoCardProps) {
  const { t } = useTranslation();
  const progressPercent =
    watchedSeconds != null && lengthSeconds > 0
      ? Math.min((watchedSeconds / lengthSeconds) * 100, 100)
      : 0;
  const watchUrl = buildWatchUrl(videoId, watchedSeconds, playlistId);

  const thumbnail = (
    <VideoThumbnail
      watchUrl={watchUrl}
      thumbnailUrl={thumbnailUrl}
      title={title}
      lengthSeconds={lengthSeconds}
      progressPercent={progressPercent}
      size={layout}
    />
  );

  if (layout === "row") {
    return (
      <article class="flex gap-4 group">
        {thumbnail}
        <div class="flex flex-col gap-3 min-w-0 flex-1 py-1">
          <a
            href={watchUrl}
            class="text-lg font-bold text-charcoal dark:text-white leading-snug line-clamp-2 no-underline hover:text-primary transition-colors"
          >
            {title}
          </a>
          <div>
            {channel && (
              <div class="flex items-center gap-2 text-sm text-text-muted-light dark:text-text-muted-dark">
                <a
                  href={`/channels/${channel.channelId}`}
                  class="flex items-center gap-1.5 no-underline text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors"
                >
                  <img
                    src={channel.iconUrl}
                    alt={channel.displayName}
                    class="size-5 rounded-full object-cover"
                  />
                  <span class="truncate">{channel.displayName}</span>
                </a>
              </div>
            )}
            {watchedAt && (
              <span class="text-xs text-text-muted-light dark:text-text-muted-dark mt-2 block">
                {t("history.watchedAgo", { time: formatTimeAgo(watchedAt, t) })}
              </span>
            )}
          </div>
        </div>
      </article>
    );
  }

  return (
    <article class="flex flex-col gap-3">
      {thumbnail}
      <div class="flex gap-3 items-start">
        {channel && (
          <a
            href={`/channels/${channel.channelId}`}
            class="size-9 rounded-full bg-gray-300 dark:bg-gray-700 flex-shrink-0 overflow-hidden cursor-pointer"
          >
            <img
              alt={channel.displayName}
              class="w-full h-full object-cover"
              src={channel.iconUrl}
            />
          </a>
        )}
        <div class="flex flex-col min-w-0 flex-1">
          <a
            href={watchUrl}
            class="text-base font-bold text-charcoal dark:text-white leading-tight line-clamp-2 cursor-pointer no-underline hover:text-primary transition-colors"
          >
            {title}
          </a>
          <div class="flex items-center justify-between text-sm text-text-muted-light dark:text-text-muted-dark mt-1">
            {channel && (
              <a
                href={`/channels/${channel.channelId}`}
                class="font-medium hover:text-charcoal dark:hover:text-white cursor-pointer truncate no-underline text-text-muted-light dark:text-text-muted-dark"
              >
                {channel.displayName}
              </a>
            )}
            {dateStr && (
              <span class={`text-xs flex-shrink-0${channel ? " ml-2" : ""}`}>
                {formatTimeAgo(dateStr, t)}
              </span>
            )}
          </div>
        </div>
      </div>
    </article>
  );
}
