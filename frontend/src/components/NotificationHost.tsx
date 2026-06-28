import { useTranslation } from "react-i18next";
import { useNotification } from "../contexts/NotificationContext";
import { Icon } from "./Icon";

export function NotificationHost() {
  const { notifications, dismiss } = useNotification();
  const { t } = useTranslation();

  if (notifications.length === 0) return null;

  return (
    <div
      class="fixed top-4 right-4 z-50 flex flex-col w-[calc(100%-2rem)] sm:w-auto sm:min-w-72 sm:max-w-sm"
      role="status"
      aria-live="polite"
    >
      {notifications.map((n) => (
        <div
          key={n.id}
          style={{
            display: "grid",
            gridTemplateRows: n.isExiting ? "0fr" : "1fr",
            paddingBottom: n.isExiting ? "0" : "0.5rem",
            transition: n.isExiting
              ? "grid-template-rows 0.2s ease-in 0.2s, padding-bottom 0.2s ease-in 0.2s"
              : "none",
            overflow: "hidden",
          }}
        >
          <div style={{ overflow: "hidden" }}>
            <div
              class="flex items-center gap-3 bg-card-light dark:bg-card-dark rounded-2xl shadow-lg ring-1 ring-border-light dark:ring-border-dark px-4 py-3"
              style={{
                animation: n.isExiting
                  ? "notification-out 0.2s ease-in forwards"
                  : "notification-in 0.25s ease-out both",
              }}
            >
              <Icon
                name={
                  n.type === "error"
                    ? "error"
                    : n.type === "success"
                      ? "check_circle"
                      : "info"
                }
                class={`text-xl shrink-0 ${
                  n.type === "error"
                    ? "text-red-500"
                    : n.type === "success"
                      ? "text-green-600 dark:text-green-400"
                      : "text-text-muted-light dark:text-text-muted-dark"
                }`}
              />
              <p class="flex-1 text-sm font-medium text-charcoal dark:text-white m-0">
                {t(n.messageKey)}
              </p>
              <button
                class="text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white bg-transparent border-none cursor-pointer p-0 shrink-0 transition-colors"
                onClick={() => dismiss(n.id)}
                aria-label={t("common.close")}
              >
                <Icon name="close" class="text-base" />
              </button>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
