import { useState, useEffect, useMemo, useCallback } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { getChannel } from "../../api/generated/channel";
import type { GetFeedChannels200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";

const CATEGORY_CODES = [
  { code: -1, key: "all", icon: "grid_view" },
  { code: 1, key: "education", icon: "school" },
  { code: 2, key: "technology", icon: "memory" },
  { code: 3, key: "economy", icon: "trending_up" },
  { code: 4, key: "politics", icon: "gavel" },
  { code: 5, key: "music", icon: "music_note" },
] as const;

const CATEGORY_KEY_MAP: Record<number, string> = {
  0: "unknown",
  1: "education",
  2: "technology",
  3: "economy",
  4: "politics",
  5: "music",
};

function getCategoryBadgeClasses(code: number): string {
  switch (code) {
    case 1:
      return "text-primary bg-primary/10";
    case 2:
      return "text-blue-600 bg-blue-100 dark:bg-blue-900/30 dark:text-blue-300";
    case 3:
      return "text-green-600 bg-green-100 dark:bg-green-900/30 dark:text-green-300";
    case 4:
      return "text-orange-600 bg-orange-100 dark:bg-orange-900/30 dark:text-orange-300";
    case 5:
      return "text-pink-600 bg-pink-100 dark:bg-pink-900/30 dark:text-pink-300";
    default:
      return "text-[#637588] bg-gray-100 dark:bg-gray-700/50 dark:text-gray-300";
  }
}

function ExploreContent() {
  const { t } = useTranslation();
  useTitle(t("explore.pageTitle"));

  const [channels, setChannels] = useState<GetFeedChannels200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState(-1);

  const loadChannels = useCallback(async () => {
    setIsLoading(true);
    setError(false);
    try {
      const res = await getChannel().getFeedChannels();
      setChannels(res.items);
    } catch {
      setError(true);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadChannels();
  }, [loadChannels]);

  const filteredChannels = useMemo(() => {
    if (selectedCategory < 0) return channels;
    return channels.filter((ch) => ch.category_code === selectedCategory);
  }, [channels, selectedCategory]);

  return (
    <DashboardLayout>
      <div class="flex-1 flex justify-center py-8 px-4 sm:px-8">
        <div class="w-full max-w-[1440px] flex flex-col gap-8">
            {/* Header */}
            <div class="flex flex-col gap-6">
              <div>
                <h1 class="text-charcoal dark:text-white text-4xl font-black leading-tight tracking-[-0.033em]">
                  {t("explore.title")}
                </h1>
              </div>

              {/* Category pills */}
              <div class="flex gap-2 overflow-x-auto pb-2 -mx-2 px-2">
                {CATEGORY_CODES.map((cat) => {
                  const isActive = cat.code === selectedCategory;
                  return (
                    <button
                      key={cat.code}
                      class={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-bold whitespace-nowrap cursor-pointer border-none transition-colors ${
                        isActive
                          ? "bg-primary text-white"
                          : "bg-card-light dark:bg-card-dark text-charcoal dark:text-gray-300 hover:bg-primary/10"
                      }`}
                      onClick={() => setSelectedCategory(cat.code)}
                    >
                      <Icon name={cat.icon} class="text-[16px]" />
                      {cat.key === "all"
                        ? t("explore.allCategories")
                        : t(`explore.categories.${cat.key}`)}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Channel Grid */}
            {isLoading ? (
              <LoadingSpinner />
            ) : error ? (
              <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
                <Icon name="error_outline" class="text-5xl mb-4" />
                <p class="text-lg font-medium">{t("explore.loadError")}</p>
                <button
                  class="mt-4 px-4 py-2 bg-primary text-white rounded-lg font-medium text-sm hover:bg-primary/90 transition-colors cursor-pointer border-none"
                  onClick={loadChannels}
                >
                  {t("explore.retry")}
                </button>
              </div>
            ) : filteredChannels.length === 0 ? (
              <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
                <Icon name="search_off" class="text-5xl mb-4" />
                <p class="text-lg font-medium">{t("explore.noChannels")}</p>
              </div>
            ) : (
              <section class="flex flex-col gap-4">
                <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-4">
                  {filteredChannels.map((ch) => {
                    const categoryKey =
                      CATEGORY_KEY_MAP[ch.category_code] ?? "unknown";

                    return (
                      <a
                        key={ch.channel_id}
                        href={`/channels/${ch.channel_id}`}
                        class="bg-card-light dark:bg-card-dark p-6 rounded-xl border border-border-light dark:border-border-dark hover:border-primary/50 dark:hover:border-primary/50 transition-colors flex flex-col gap-4 no-underline min-h-[180px]"
                      >
                        <div class="flex items-center gap-3">
                          <img
                            alt={ch.external_channel_display_name}
                            loading="lazy"
                            class="size-12 rounded-full object-cover bg-gray-100"
                            src={ch.external_channel_icon_url}
                          />
                          <div>
                            <p class="font-bold text-charcoal dark:text-white text-base leading-snug">
                              {ch.external_channel_display_name}
                            </p>
                            <span
                              class={`text-xs font-medium px-2 py-0.5 rounded-full ${getCategoryBadgeClasses(ch.category_code)}`}
                            >
                              {t(`explore.categories.${categoryKey}`)}
                            </span>
                          </div>
                        </div>
                        <p class="text-text-muted-light dark:text-text-muted-dark text-sm leading-relaxed line-clamp-3">
                          {ch.valuable_description}
                        </p>
                      </a>
                    );
                  })}
                </div>
              </section>
            )}
        </div>
      </div>
    </DashboardLayout>
  );
}

export default function Explore() {
  return (
    <ProtectedRoute>
      <ExploreContent />
    </ProtectedRoute>
  );
}
