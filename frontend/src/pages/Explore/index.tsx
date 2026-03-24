import { useState, useEffect, useMemo, useCallback } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import { getChannel } from "../../api/generated/channel";
import type { GetFeedChannels200ItemsItem } from "../../api/generated/antiYtApi.schemas";

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
  const [searchQuery, setSearchQuery] = useState("");
  const [subscribingIds, setSubscribingIds] = useState<Set<string>>(new Set());
  const [subscribedIds, setSubscribedIds] = useState<Set<string>>(new Set());

  const loadChannels = useCallback(async () => {
    setIsLoading(true);
    setError(false);
    try {
      const [feedRes, subsRes] = await Promise.allSettled([
        getChannel().getFeedChannels(),
        getChannel().getChannelsSubscribed({ limit: 50 }),
      ]);
      if (feedRes.status === "fulfilled") {
        setChannels(feedRes.value.items);
      } else {
        setError(true);
      }
      if (subsRes.status === "fulfilled") {
        setSubscribedIds(
          new Set(subsRes.value.items.map((s) => s.channel_id)),
        );
      }
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
    let result = channels;
    if (selectedCategory >= 0) {
      result = result.filter((ch) => ch.category_code === selectedCategory);
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      result = result.filter(
        (ch) =>
          ch.external_channel_display_name.toLowerCase().includes(q) ||
          ch.valuable_description.toLowerCase().includes(q),
      );
    }
    return result;
  }, [channels, selectedCategory, searchQuery]);

  const handleSubscribe = async (channelCustomUrl: string, channelId: string) => {
    if (subscribedIds.has(channelId) || subscribingIds.has(channelId)) return;
    setSubscribingIds((prev) => new Set(prev).add(channelId));
    try {
      await getChannel().postChannelsSubscribe({
        channel_id: channelCustomUrl,
      });
      setSubscribedIds((prev) => new Set(prev).add(channelId));
    } catch {
      // ignore
    } finally {
      setSubscribingIds((prev) => {
        const next = new Set(prev);
        next.delete(channelId);
        return next;
      });
    }
  };

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
                  const isActive =
                    cat.code === selectedCategory ||
                    (cat.code === -1 && selectedCategory === -1);
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
                      <span class="material-symbols-outlined text-[16px]">
                        {cat.icon}
                      </span>
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
                <span class="material-symbols-outlined text-5xl mb-4">
                  error_outline
                </span>
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
                <span class="material-symbols-outlined text-5xl mb-4">
                  search_off
                </span>
                <p class="text-lg font-medium">{t("explore.noChannels")}</p>
              </div>
            ) : (
              <section class="flex flex-col gap-4">
                <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-4">
                  {filteredChannels.map((ch) => {
                    const isSubscribed = subscribedIds.has(ch.channel_id);
                    const isSubscribing = subscribingIds.has(ch.channel_id);
                    const categoryKey =
                      CATEGORY_KEY_MAP[ch.category_code] ?? "unknown";

                    return (
                      <div
                        key={ch.channel_id}
                        class="bg-card-light dark:bg-card-dark p-5 rounded-xl border border-border-light dark:border-border-dark hover:border-primary/50 dark:hover:border-primary/50 transition-colors flex flex-col gap-4 group"
                      >
                        <div class="flex items-start justify-between">
                          <div class="flex items-center gap-3">
                            <img
                              alt={ch.external_channel_display_name}
                              class="size-12 rounded-full object-cover bg-gray-100"
                              src={ch.external_channel_icon_url}
                            />
                            <div>
                              <h3 class="font-bold text-charcoal dark:text-white text-base leading-snug">
                                {ch.external_channel_display_name}
                              </h3>
                              <span
                                class={`text-xs font-medium px-2 py-0.5 rounded-full ${getCategoryBadgeClasses(ch.category_code)}`}
                              >
                                {t(`explore.categories.${categoryKey}`)}
                              </span>
                            </div>
                          </div>
                        </div>
                        <p class="text-text-muted-light dark:text-text-muted-dark text-sm leading-relaxed line-clamp-2">
                          {ch.valuable_description}
                        </p>
                        {isSubscribed ? (
                          <button class="mt-auto w-full py-2 px-4 rounded-lg bg-primary/10 dark:bg-primary/20 text-primary cursor-default text-sm font-bold flex items-center justify-center gap-2 border-none">
                            <span class="material-symbols-outlined text-[18px]">
                              check_circle
                            </span>
                            {t("explore.added")}
                          </button>
                        ) : (
                          <button
                            class="mt-auto w-full py-2 px-4 rounded-lg border border-primary text-primary hover:bg-primary hover:text-white transition-all text-sm font-bold flex items-center justify-center gap-2 cursor-pointer bg-transparent disabled:opacity-50"
                            onClick={() =>
                              handleSubscribe(
                                ch.external_channel_cusom_url,
                                ch.channel_id,
                              )
                            }
                            disabled={isSubscribing}
                          >
                            <span class="material-symbols-outlined text-[18px]">
                              playlist_add
                            </span>
                            {isSubscribing
                              ? t("explore.adding")
                              : t("explore.addToWhitelist")}
                          </button>
                        )}
                      </div>
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
