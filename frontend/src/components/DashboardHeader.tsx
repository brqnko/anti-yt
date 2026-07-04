import { useState, useMemo } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { Icon } from "./Icon";
import { AuthPromptDialog } from "./AuthPromptDialog";
import { SearchFilterDialog, parseSearchOrder, parseSearchType, type SearchFilters } from "./SearchFilterDialog";
import { useAuth } from "../contexts/AuthContext";

function parseFiltersFromURL(qs: string): SearchFilters {
  const p = new URLSearchParams(qs);
  const filters: SearchFilters = {};
  const order = parseSearchOrder(p.get("order"));
  if (order) filters.order = order;
  if (p.get("published_after")) filters.published_after = p.get("published_after")!;
  if (p.get("published_before")) filters.published_before = p.get("published_before")!;
  if (p.get("region_code")) filters.region_code = p.get("region_code")!;
  if (p.get("relevance_language")) filters.relevance_language = p.get("relevance_language")!;
  const type = parseSearchType(p.get("type"));
  if (type) filters.type = type;
  return filters;
}

function buildSearchURL(query: string, filters: SearchFilters): string {
  const p = new URLSearchParams();
  p.set("q", query);
  if (filters.order) p.set("order", filters.order);
  if (filters.published_after) p.set("published_after", filters.published_after);
  if (filters.published_before) p.set("published_before", filters.published_before);
  if (filters.region_code) p.set("region_code", filters.region_code);
  if (filters.relevance_language) p.set("relevance_language", filters.relevance_language);
  if (filters.type) p.set("type", filters.type);
  return `/search?${p.toString()}`;
}

export function DashboardHeader({
  sidebarOpen = false,
  onToggleSidebar,
}: {
  sidebarOpen?: boolean;
  onToggleSidebar?: () => void;
}) {
  const { t } = useTranslation();
  const { url, route } = useLocation();
  const { isAuthenticated, isLoading: isAuthLoading } = useAuth();
  const [showAuthPrompt, setShowAuthPrompt] = useState(false);
  const qs = url.split("?")[1] || "";
  const initialQuery = url.startsWith("/search")
    ? new URLSearchParams(qs).get("q") || ""
    : "";
  const [searchInput, setSearchInput] = useState(initialQuery);
  const [filterOpen, setFilterOpen] = useState(false);

  const currentFilters = useMemo(() => parseFiltersFromURL(qs), [qs]);
  const activeFilterCount = useMemo(
    () => Object.values(currentFilters).filter(Boolean).length,
    [currentFilters],
  );

  const handleSearch = (e: Event) => {
    e.preventDefault();
    if (isAuthLoading) return;
    if (!isAuthenticated) { setShowAuthPrompt(true); return; }
    (document.activeElement as HTMLElement | null)?.blur();
    const q = searchInput.trim();
    if (q) {
      route(buildSearchURL(q, currentFilters));
    }
  };

  const handleApplyFilters = (filters: SearchFilters) => {
    const q = searchInput.trim();
    if (q) {
      route(buildSearchURL(q, filters));
    }
  };

  return (
    <header class="sticky top-0 z-50 flex items-center justify-between border-b border-solid border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark px-6 py-3 gap-4 overflow-hidden">
      <div class="flex items-center gap-3 shrink-0">
        {onToggleSidebar && (
          <button
            class="hidden tablet:inline-flex p-1.5 rounded-lg hover:bg-black/5 dark:hover:bg-white/5 bg-transparent border-none cursor-pointer text-charcoal dark:text-white"
            onClick={onToggleSidebar}
            aria-label={t("dashboard.toggleSidebar")}
            aria-expanded={sidebarOpen}
          >
            <Icon name="menu" class="text-2xl" />
          </button>
        )}
        <a href="/" class="no-underline text-charcoal dark:text-white">
          <span class="text-xl font-bold tracking-tight whitespace-nowrap">anti-yt</span>
        </a>
      </div>

      <form
        class="flex-1 min-w-0 max-w-xl mx-auto flex items-center"
        onSubmit={handleSearch}
        role="search"
      >
        <button type="submit" class="hidden" tabIndex={-1} aria-hidden="true" />
        <div class="flex flex-1 min-w-0 items-center rounded-full border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark overflow-hidden focus-within:border-primary transition-colors">
          <input
            type="text"
            value={searchInput}
            onInput={(e) => setSearchInput((e.target as HTMLInputElement).value)}
            placeholder={t("search.inputPlaceholder")}
            class="flex-1 bg-transparent px-4 py-2 text-base sm:text-sm text-charcoal dark:text-white outline-none placeholder:text-text-muted-light dark:placeholder:text-text-muted-dark"
            aria-label={t("search.inputPlaceholder")}
          />
          {searchInput && (
            <button
              type="button"
              class="shrink-0 p-1.5 rounded-full hover:bg-black/5 dark:hover:bg-white/5 bg-transparent border-none cursor-pointer text-text-muted-light dark:text-text-muted-dark"
              onClick={() => setSearchInput("")}
              aria-label={t("search.clear")}
            >
              <Icon name="close" class="text-xl" />
            </button>
          )}
          <button
            type="button"
            class="relative shrink-0 p-1.5 mr-1.5 rounded-full hover:bg-black/5 dark:hover:bg-white/5 bg-transparent border-none cursor-pointer text-charcoal dark:text-white"
            onClick={() => setFilterOpen(true)}
            aria-label={t("search.filters.title")}
          >
            <Icon name="tune" class="text-xl" />
            {activeFilterCount > 0 && (
              <span class="absolute -top-0.5 -right-0.5 size-4 flex items-center justify-center rounded-full bg-primary text-white text-[10px] font-bold leading-none">
                {activeFilterCount}
              </span>
            )}
          </button>
        </div>
      </form>

      <SearchFilterDialog
        open={filterOpen}
        onClose={() => setFilterOpen(false)}
        filters={currentFilters}
        onApply={handleApplyFilters}
      />

      <div class="hidden tablet:flex items-center gap-4 shrink-0">
        <a
          href="/profile"
          onClick={(e) => {
            if (!isAuthLoading && !isAuthenticated) {
              e.preventDefault();
              e.stopPropagation();
              setShowAuthPrompt(true);
            }
          }}
          class="size-9 flex items-center justify-center rounded-full bg-primary/10 ring-2 ring-primary/20 cursor-pointer text-primary no-underline"
          aria-label={t("profile.pageTitle")}
        >
          <Icon name="person" class="text-[20px]" />
        </a>
      </div>
      <AuthPromptDialog open={showAuthPrompt} onClose={() => setShowAuthPrompt(false)} />
    </header>
  );
}
