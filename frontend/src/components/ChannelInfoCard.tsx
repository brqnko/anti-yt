import { useState, useEffect, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { formatSubscriberCount } from "../utils/format";
import { getChannel } from "../api/generated/channel";
import type { GetChannelsChannelId200 } from "../api/generated/antiYtApi.schemas";
import { Linkify } from "./Linkify";
import { Icon } from "./Icon";
import { Toggle } from "./Toggle";

function ExpandableDescription({ description }: { description: string }) {
  const { t } = useTranslation();
  const [expanded, setExpanded] = useState(false);
  const [clamped, setClamped] = useState(false);
  const ref = useRef<HTMLParagraphElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (el) setClamped(el.scrollHeight > el.clientHeight);
  }, [description]);

  return (
    <>
      <div class="h-px bg-border-light dark:bg-border-dark my-5" />
      <p
        ref={ref}
        class={`text-sm text-text-muted-light dark:text-text-muted-dark leading-relaxed whitespace-pre-line ${expanded ? "" : "line-clamp-3"}`}
      >
        <Linkify text={description} />
      </p>
      {clamped && (
        <button
          type="button"
          class="text-sm font-medium text-primary hover:text-primary/80 transition-colors bg-transparent border-none cursor-pointer p-0 mt-2"
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? t("channelDetail.showLess") : t("channelDetail.showMore")}
        </button>
      )}
    </>
  );
}

export function ChannelInfoCard({
  channelInfo,
  isSubscribed,
  onToggleSubscription,
  isToggling,
}: {
  channelInfo: GetChannelsChannelId200;
  isSubscribed: boolean;
  onToggleSubscription: () => void;
  isToggling: boolean;
}) {
  const { t } = useTranslation();

  return (
    <div class="bg-card-light dark:bg-card-dark rounded-xl border border-border-light dark:border-border-dark mb-8 p-6">
      <div class="flex flex-row gap-4 md:gap-6 items-start md:items-center">
        {/* Avatar */}
        <div class="shrink-0">
          <div class="size-16 md:size-28 rounded-full bg-gray-200 dark:bg-gray-800 overflow-hidden border-2 border-border-light dark:border-border-dark">
            <img
              src={channelInfo.external_channel_icon_url}
              alt={channelInfo.external_channel_display_name}
              class="w-full h-full object-cover"
            />
          </div>
        </div>

        {/* Channel Text Info + Whitelist Toggle */}
        <div class="flex-1 min-w-0 flex flex-col md:flex-row md:items-center gap-3">
          {/* Channel Text Info */}
          <div class="flex-1 min-w-0">
            <h1 class="text-xl md:text-3xl font-bold mb-1 truncate">{channelInfo.external_channel_display_name}</h1>
            <div class="flex flex-wrap gap-x-4 gap-y-1 text-sm text-text-muted-light dark:text-text-muted-dark">
              <span>
                {formatSubscriberCount(channelInfo.external_channel_subscribers_count)} {t("channelDetail.subscribers")}
              </span>
              {channelInfo.external_channel_custom_id && (
                <span>
                  {channelInfo.external_channel_custom_id}
                </span>
              )}
              <a
                href={`https://www.youtube.com/${channelInfo.external_channel_custom_id}`}
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-1 text-text-muted-light dark:text-text-muted-dark hover:text-primary transition-colors no-underline"
              >
                <Icon name="open_in_new" class="text-sm" />
                {t("channelDetail.openOnYouTube")}
              </a>
            </div>
          </div>

          {/* Whitelist Toggle */}
          <div class="flex-shrink-0">
            <div class="bg-background-light dark:bg-background-dark border border-primary/20 p-4 rounded-xl flex items-center gap-4">
              <div class="flex flex-col">
                <span class="text-sm font-bold">{t("channelDetail.whitelistChannel")}</span>
              </div>
              <Toggle checked={isSubscribed} disabled={isToggling} onClick={onToggleSubscription} />
            </div>
          </div>
        </div>
      </div>

      {/* Description */}
      {channelInfo.external_channel_description && (
        <ExpandableDescription description={channelInfo.external_channel_description} />
      )}
    </div>
  );
}
