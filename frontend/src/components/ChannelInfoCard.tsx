import { useState, useEffect, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { formatSubscriberCount } from "../utils/format";
import type { GetChannelsChannelId200 } from "../api/generated/antiYtApi.schemas";
import { Linkify } from "./Linkify";
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
      <div class="mt-4 max-w-3xl">
        <p
          ref={ref}
          class={`text-sm text-text-muted-light dark:text-text-muted-dark leading-relaxed whitespace-pre-line ${expanded ? "" : "line-clamp-3"}`}
        >
          <Linkify text={description} />
        </p>
      </div>
      {clamped && (
        <button
          type="button"
          class="text-sm font-medium text-primary hover:text-primary/80 bg-transparent border-none cursor-pointer p-0 mt-2"
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
    <section class="mb-8">
      <div class="flex flex-col sm:flex-row gap-5 md:gap-6 items-start">
        <div class="size-20 md:size-24 shrink-0 rounded-full bg-gray-200 dark:bg-gray-800 overflow-hidden ring-1 ring-border-light dark:ring-border-dark">
          <img
            src={channelInfo.external_channel_icon_url}
            alt={channelInfo.external_channel_display_name}
            class="w-full h-full object-cover"
          />
        </div>

        <div class="flex-1 min-w-0 w-full">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="min-w-0">
              <h1 class="text-2xl md:text-3xl font-bold leading-tight text-charcoal dark:text-white break-words">
                {channelInfo.external_channel_display_name}
              </h1>
              <div class="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-sm text-text-muted-light dark:text-text-muted-dark">
                <span>
                  {formatSubscriberCount(channelInfo.external_channel_subscribers_count)} {t("channelDetail.subscribers")}
                </span>
                {channelInfo.external_channel_custom_id && (
                  <>
                    <span aria-hidden="true">·</span>
                    <span class="break-all">
                      {channelInfo.external_channel_custom_id}
                    </span>
                  </>
                )}
              </div>
            </div>

            <div class="flex flex-shrink-0 items-center gap-3">
              <span class="text-sm font-bold text-charcoal dark:text-white">
                {t("channelDetail.whitelistChannel")}
              </span>
              <Toggle checked={isSubscribed} disabled={isToggling} onClick={onToggleSubscription} />
            </div>
          </div>

          {channelInfo.external_channel_description && (
            <ExpandableDescription description={channelInfo.external_channel_description} />
          )}
        </div>
      </div>
    </section>
  );
}
