import { useState, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { Icon } from "./Icon";
import { GetSearchOrder } from "../api/generated/antiYtApi.schemas";

export interface SearchFilters {
  order?: string;
  published_after?: string;
  published_before?: string;
  region_code?: string;
  relevance_language?: string;
}

interface SearchFilterDialogProps {
  open: boolean;
  onClose: () => void;
  filters: SearchFilters;
  onApply: (filters: SearchFilters) => void;
}

const ORDER_OPTIONS = Object.values(GetSearchOrder);

const REGION_CODES = ["", "JP", "US"] as const;
const LANGUAGES = ["", "ja", "en"] as const;

export function SearchFilterDialog({
  open,
  onClose,
  filters,
  onApply,
}: SearchFilterDialogProps) {
  const { t } = useTranslation();
  const [draft, setDraft] = useState<SearchFilters>(filters);

  useEffect(() => {
    if (open) setDraft(filters);
  }, [open, filters]);

  useEffect(() => {
    if (!open) return;
    document.body.style.overflow = "hidden";
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      document.body.style.overflow = "";
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [open, onClose]);

  if (!open) return null;

  const handleApply = () => {
    onApply(draft);
    onClose();
  };

  const handleReset = () => {
    const empty: SearchFilters = {};
    setDraft(empty);
    onApply(empty);
    onClose();
  };

  const hasFilters =
    draft.order || draft.published_after || draft.published_before || draft.region_code || draft.relevance_language;

  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      aria-label={t("search.filters.title")}
    >
      <div class="absolute inset-0 bg-black/60" onClick={onClose} />
      <div class="relative bg-white dark:bg-[#2a2721] rounded-2xl ring-1 ring-black/10 dark:ring-white/10 border border-gray-100 dark:border-neutral-800 p-8 max-w-md w-full max-h-[85vh] overflow-y-auto">
        <button
          class="absolute top-4 right-4 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors bg-transparent border-none cursor-pointer"
          onClick={onClose}
          aria-label={t("search.filters.close")}
        >
          <Icon name="close" />
        </button>

        <h2 class="text-2xl font-bold text-charcoal dark:text-white mb-6">
          {t("search.filters.title")}
        </h2>

        <div class="space-y-5">
          {/* Sort order */}
          <div>
            <label class="block text-sm font-medium text-charcoal dark:text-white mb-1.5">
              {t("search.filters.order")}
            </label>
            <select
              class="w-full px-3 py-2.5 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all text-sm"
              value={draft.order || ""}
              onChange={(e) =>
                setDraft({ ...draft, order: (e.target as HTMLSelectElement).value || undefined })
              }
            >
              <option value="">{t("search.filters.orderDefault")}</option>
              {ORDER_OPTIONS.map((o) => (
                <option key={o} value={o}>
                  {t(`search.filters.orderOptions.${o}`)}
                </option>
              ))}
            </select>
          </div>

          {/* Published after */}
          <div>
            <label class="block text-sm font-medium text-charcoal dark:text-white mb-1.5">
              {t("search.filters.publishedAfter")}
            </label>
            <input
              type="date"
              class="w-full px-3 py-2.5 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all text-sm"
              value={draft.published_after || ""}
              onInput={(e) =>
                setDraft({ ...draft, published_after: (e.target as HTMLInputElement).value || undefined })
              }
            />
          </div>

          {/* Published before */}
          <div>
            <label class="block text-sm font-medium text-charcoal dark:text-white mb-1.5">
              {t("search.filters.publishedBefore")}
            </label>
            <input
              type="date"
              class="w-full px-3 py-2.5 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all text-sm"
              value={draft.published_before || ""}
              onInput={(e) =>
                setDraft({ ...draft, published_before: (e.target as HTMLInputElement).value || undefined })
              }
            />
          </div>

          {/* Region code */}
          <div>
            <label class="block text-sm font-medium text-charcoal dark:text-white mb-1.5">
              {t("search.filters.regionCode")}
            </label>
            <select
              class="w-full px-3 py-2.5 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all text-sm"
              value={draft.region_code || ""}
              onChange={(e) =>
                setDraft({ ...draft, region_code: (e.target as HTMLSelectElement).value || undefined })
              }
            >
              <option value="">{t("search.filters.regionDefault")}</option>
              {REGION_CODES.filter(Boolean).map((code) => (
                <option key={code} value={code}>
                  {t(`search.filters.regions.${code}`)}
                </option>
              ))}
            </select>
          </div>

          {/* Relevance language */}
          <div>
            <label class="block text-sm font-medium text-charcoal dark:text-white mb-1.5">
              {t("search.filters.relevanceLanguage")}
            </label>
            <select
              class="w-full px-3 py-2.5 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all text-sm"
              value={draft.relevance_language || ""}
              onChange={(e) =>
                setDraft({ ...draft, relevance_language: (e.target as HTMLSelectElement).value || undefined })
              }
            >
              <option value="">{t("search.filters.languageDefault")}</option>
              {LANGUAGES.filter(Boolean).map((lang) => (
                <option key={lang} value={lang}>
                  {t(`search.filters.languages.${lang}`)}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Actions */}
        <div class="flex justify-between mt-8">
          <button
            class="px-4 py-2 rounded-xl text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors bg-transparent border-none cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
            onClick={handleReset}
            disabled={!hasFilters}
          >
            {t("search.filters.reset")}
          </button>
          <div class="flex gap-3">
            <button
              class="px-4 py-2 rounded-xl text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors bg-transparent border-none cursor-pointer"
              onClick={onClose}
            >
              {t("search.filters.cancel")}
            </button>
            <button
              class="px-4 py-2 rounded-xl text-sm font-bold text-white bg-primary hover:bg-primary/90 transition-colors border-none cursor-pointer"
              onClick={handleApply}
            >
              {t("search.filters.apply")}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
