import { useTranslation } from "react-i18next";
import { Icon } from "./Icon";

export function ExploreChannelsBanner() {
  const { t } = useTranslation();

  return (
    <a
      href="/channels/explore"
      class="flex items-center gap-4 px-5 py-14 rounded-xl no-underline overflow-hidden relative"
    >
      <div
        class="absolute inset-0"
        style={{ backgroundImage: "url('/explore-banner.webp')", backgroundSize: "cover", backgroundPosition: "center 70%" }}
      />
      <div class="absolute inset-0 bg-black/50" />
      <div class="flex-1 min-w-0 relative z-10">
        <span class="text-xl font-bold text-white">
          {t("channels.exploreLink")}
        </span>
        <p class="text-base text-white/80 mt-0.5">
          {t("channels.exploreLinkDesc")}
        </p>
      </div>
      <Icon name="chevron_right" class="text-xl text-white/80 shrink-0 relative z-10" />
    </a>
  );
}
