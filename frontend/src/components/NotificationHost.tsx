import { useTranslation } from "react-i18next";
import { useNotification } from "../contexts/NotificationContext";
import { Icon } from "./Icon";

export function NotificationHost() {
  const { notifications, dismiss } = useNotification();
  const { t } = useTranslation();

  if (notifications.length === 0) return null;

  return (
    <div
      class="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-sm w-[calc(100%-2rem)] sm:w-96"
      role="status"
      aria-live="polite"
    >
      {notifications.map((n) => (
        <div
          key={n.id}
          class="flex items-start gap-3 bg-white dark:bg-[#2a2721] rounded-xl ring-1 ring-black/10 dark:ring-white/10 border border-gray-100 dark:border-neutral-800 shadow-lg px-4 py-3"
        >
          <Icon
            name={n.type === "error" ? "error" : "bolt"}
            class={`text-xl shrink-0 mt-0.5 ${n.type === "error" ? "text-red-500" : "text-text-muted-light dark:text-text-muted-dark"}`}
          />
          <p class="flex-1 text-sm text-charcoal dark:text-white m-0">
            {t(n.messageKey)}
          </p>
          <button
            class="text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors bg-transparent border-none cursor-pointer p-0 shrink-0"
            onClick={() => dismiss(n.id)}
            aria-label={t("common.close")}
          >
            <Icon name="close" />
          </button>
        </div>
      ))}
    </div>
  );
}
