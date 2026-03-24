import { useState } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { Logo } from "./Logo";

export function DashboardHeader({
  sidebarOpen = false,
  onToggleSidebar,
}: {
  sidebarOpen?: boolean;
  onToggleSidebar?: () => void;
}) {
  const { t } = useTranslation();
  const { url, route } = useLocation();
  const initialQuery = url.startsWith("/search")
    ? new URLSearchParams(url.split("?")[1] || "").get("q") || ""
    : "";
  const [searchInput, setSearchInput] = useState(initialQuery);

  const handleSearch = (e: Event) => {
    e.preventDefault();
    const q = searchInput.trim();
    if (q) {
      route(`/search?q=${encodeURIComponent(q)}`);
    }
  };

  return (
    <header class="sticky top-0 z-50 flex items-center justify-between whitespace-nowrap border-b border-solid border-border-light dark:border-border-dark bg-background-light/95 dark:bg-background-dark/95 backdrop-blur-md px-6 py-3 gap-4">
      <div class="flex items-center gap-3 shrink-0">
        {onToggleSidebar && (
          <button
            class="p-1.5 rounded-lg hover:bg-black/5 dark:hover:bg-white/5 transition-colors bg-transparent border-none cursor-pointer text-charcoal dark:text-white"
            onClick={onToggleSidebar}
            aria-label={t("dashboard.toggleSidebar")}
            aria-expanded={sidebarOpen}
          >
            <span class="material-symbols-outlined text-2xl">menu</span>
          </button>
        )}
        <Logo />
      </div>

      <form
        class="flex-1 max-w-xl mx-auto flex"
        onSubmit={handleSearch}
        role="search"
      >
        <div class="flex w-full rounded-full border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark overflow-hidden focus-within:border-primary transition-colors">
          <input
            type="text"
            value={searchInput}
            onInput={(e) => setSearchInput((e.target as HTMLInputElement).value)}
            placeholder={t("search.inputPlaceholder")}
            class="flex-1 bg-transparent px-4 py-2 text-sm text-charcoal dark:text-white outline-none placeholder:text-text-muted-light dark:placeholder:text-text-muted-dark"
            aria-label={t("search.inputPlaceholder")}
          />
          <button
            type="submit"
            class="px-4 py-2 bg-transparent border-none cursor-pointer text-text-muted-light dark:text-text-muted-dark hover:text-primary transition-colors"
            aria-label={t("search.pageTitle")}
          >
            <span class="material-symbols-outlined text-[20px]">search</span>
          </button>
        </div>
      </form>

      <div class="flex items-center gap-4 shrink-0">
        <
          href="/profile"
          class="size-9 flex items-center justify-center rounded-full bg-primary/10 ring-2 ring-primary/20 cursor-pointer text-primary no-underline"
          aria-label={t("profile.pageTitle")}
        >
          <span class="material-symbols-outlined text-[20px]">person</span>
        </a>
      </div>
    </header>
  );
}
