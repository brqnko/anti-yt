import { useState, useEffect, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { formatDuration, formatTimeAgo } from "../utils/format";
import { buildWatchUrl } from "../utils/url";
import { Icon } from "./Icon";

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
  isSubscribed?: boolean;
  onToggleSubscription?: () => Promise<void>;
  isWatched?: boolean;
  onMarkWatched?: () => Promise<void>;
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
        size === "card" ? "rounded-xl" : "rounded-xl sm:w-48 sm:flex-shrink-0 sm:rounded-lg md:w-60"
      }`}
    >
      <img
        src={thumbnailUrl}
        alt={title}
        loading="lazy"
        class="absolute inset-0 w-full h-full object-cover"
      />
      <div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover/thumb:opacity-100 transition-opacity duration-300" />
      <span class="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-bold px-1.5 py-0.5 rounded">
        {formatDuration(lengthSeconds)}
      </span>
      <div class="absolute inset-0 flex items-center justify-center opacity-0 group-hover/thumb:opacity-100 transition-opacity duration-300 pointer-events-none">
        <div
          class={`rounded-full bg-primary/90 flex items-center justify-center text-white ${
            size === "card" ? "size-12" : "size-10"
          }`}
        >
          <Icon
            name="play_arrow"
            class={size === "card" ? "text-[28px] ml-1" : "text-[22px] ml-0.5"}
          />
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

function VideoCardMenu({
  isSubscribed,
  onToggleSubscription,
  isWatched,
  onMarkWatched,
}: {
  isSubscribed?: boolean;
  onToggleSubscription?: () => Promise<void>;
  isWatched?: boolean;
  onMarkWatched?: () => Promise<void>;
}) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [toggling, setToggling] = useState(false);
  const [marking, setMarking] = useState(false);
  const [subscribed, setSubscribed] = useState(isSubscribed);
  const dialogRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setSubscribed(isSubscribed);
  }, [isSubscribed]);

  useEffect(() => {
    if (!open) return;
    document.body.style.overflow = "hidden";
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      document.body.style.overflow = "";
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  const handleToggle = async () => {
    if (toggling) return;
    setToggling(true);
    try {
      await onToggleSubscription();
      setSubscribed((prev) => !prev);
    } finally {
      setToggling(false);
    }
  };

  const handleMarkWatched = async () => {
    if (marking || !onMarkWatched) return;
    setMarking(true);
    try {
      await onMarkWatched();
      setOpen(false);
    } finally {
      setMarking(false);
    }
  };

  return (
    <>
      <button
        type="button"
        class="flex-shrink-0 p-0.5 -mr-1 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors bg-transparent border-none cursor-pointer text-text-muted-light dark:text-text-muted-dark"
        onClick={(e) => {
          e.preventDefault();
          e.stopPropagation();
          setOpen(true);
        }}
        aria-label={t("videoCard.moreOptions")}
      >
        <Icon name="more_vert" class="text-[20px]" />
      </button>
      {open && (
        <div
          class="fixed inset-0 z-50 flex items-center justify-center p-4"
          role="dialog"
          aria-modal="true"
          aria-label={t("videoCard.moreOptions")}
        >
          <div
            class="absolute inset-0 bg-black/60"
            onClick={() => setOpen(false)}
          />
          <div
            ref={dialogRef}
            class="relative bg-white dark:bg-[#2a2721] rounded-2xl ring-1 ring-black/10 dark:ring-white/10 border border-gray-100 dark:border-neutral-800 p-8 max-w-sm w-full"
          >
            <button
              class="absolute top-4 right-4 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors bg-transparent border-none cursor-pointer"
              onClick={() => setOpen(false)}
              aria-label={t("common.close")}
            >
              <Icon name="close" />
            </button>
            {onToggleSubscription && (
              <div class="flex items-center justify-between">
                <div class="flex flex-col">
                  <span class="text-sm font-bold text-charcoal dark:text-white">
                    {t("videoCard.subscribeChannel")}
                  </span>
                  <span class="text-xs text-text-muted-light dark:text-text-muted-dark">
                    {t("videoCard.subscribeDesc")}
                  </span>
                </div>
                <button
                  class="relative inline-flex items-center cursor-pointer bg-transparent border-none p-0 flex-shrink-0 ml-3"
                  onClick={handleToggle}
                  disabled={toggling}
                >
                  <div
                    class={`w-14 h-7 rounded-full transition-colors duration-200 ${
                      subscribed ? "bg-primary" : "bg-gray-200 dark:bg-gray-700"
                    } ${toggling ? "opacity-50" : ""}`}
                  >
                    <div
                      class={`absolute top-0.5 left-[4px] bg-white border border-gray-300 rounded-full h-6 w-6 transition-transform duration-200 ${
                        subscribed ? "translate-x-full" : ""
                      }`}
                    />
                  </div>
                </button>
              </div>
            )}
            {onMarkWatched && (
              <div class={`flex items-center justify-between${onToggleSubscription ? " mt-4" : ""}`}>
                <div class="flex flex-col">
                  <span class="text-sm font-bold text-charcoal dark:text-white">
                    {isWatched ? t("videoCard.unmarkWatched") : t("videoCard.markWatched")}
                  </span>
                  <span class="text-xs text-text-muted-light dark:text-text-muted-dark">
                    {isWatched ? t("videoCard.unmarkWatchedDesc") : t("videoCard.markWatchedDesc")}
                  </span>
                </div>
                <button
                  class={`flex-shrink-0 ml-3 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                    marking
                      ? "bg-gray-200 dark:bg-gray-700 text-text-muted-light dark:text-text-muted-dark cursor-not-allowed"
                      : isWatched
                        ? "bg-gray-200 dark:bg-gray-700 text-charcoal dark:text-white hover:bg-gray-300 dark:hover:bg-gray-600 cursor-pointer"
                        : "bg-primary text-white hover:bg-primary/90 cursor-pointer"
                  } border-none`}
                  onClick={handleMarkWatched}
                  disabled={marking}
                >
                  {isWatched ? t("videoCard.unmarkWatchedButton") : t("videoCard.markWatchedButton")}
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </>
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
  isSubscribed,
  onToggleSubscription,
  isWatched,
  onMarkWatched,
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
      <article class="flex flex-col sm:flex-row gap-3 sm:gap-4 group">
        {thumbnail}
        <div class="flex flex-col gap-2 sm:gap-3 min-w-0 flex-1 sm:py-1">
          <a
            href={watchUrl}
            class="text-base sm:text-lg font-bold text-charcoal dark:text-white leading-snug line-clamp-2 no-underline hover:text-primary transition-colors"
          >
            {title}
          </a>
          <div>
            {channel && (
              <div class="flex items-center gap-2 text-sm text-text-muted-light dark:text-text-muted-dark">
                <a
                  href={`/channels/${channel.channelId}`}
                  class="flex items-center gap-1.5 min-w-0 overflow-hidden no-underline text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors"
                >
                  <img
                    src={channel.iconUrl}
                    alt={channel.displayName}
                    loading="lazy"
                    class="size-5 rounded-full object-cover"
                  />
                  <span class="truncate">{channel.displayName}</span>
                </a>
                {dateStr && (
                  <span class="text-xs flex-shrink-0">
                    {formatTimeAgo(dateStr, t)}
                  </span>
                )}
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
              loading="lazy"
              class="w-full h-full object-cover"
              src={channel.iconUrl}
            />
          </a>
        )}
        <div class="flex flex-col min-w-0 flex-1">
          <div class="flex items-start gap-1">
            <a
              href={watchUrl}
              class="text-base font-bold text-charcoal dark:text-white leading-tight line-clamp-2 cursor-pointer no-underline hover:text-primary transition-colors flex-1 min-w-0"
            >
              {title}
            </a>
            {(onMarkWatched || (channel && onToggleSubscription)) && (
              <VideoCardMenu
                isSubscribed={isSubscribed ?? false}
                onToggleSubscription={onToggleSubscription}
                isWatched={isWatched}
                onMarkWatched={onMarkWatched}
              />
            )}
          </div>
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
