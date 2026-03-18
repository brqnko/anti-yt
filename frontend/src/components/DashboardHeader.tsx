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

  return (
    <header class="sticky top-0 z-50 flex items-center justify-between whitespace-nowrap border-b border-solid border-border-light dark:border-border-dark bg-background-light/95 dark:bg-background-dark/95 backdrop-blur-md px-6 py-3">
      <div class="flex items-center gap-3">
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

      <div class="flex items-center gap-4">
        <a
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
