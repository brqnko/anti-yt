import { useState, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { mutate } from "swr";
import { getChannel } from "../api/generated/channel";
import { CACHE_KEYS } from "../api/cache-keys";
import { getApiErrorCode } from "../utils/api-error";

interface AddChannelDialogProps {
  open: boolean;
  onClose: () => void;
  onAdded: () => void;
}

export function AddChannelDialog({
  open,
  onClose,
  onAdded,
}: AddChannelDialogProps) {
  const { t } = useTranslation();
  const [channelId, setChannelId] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) {
      setChannelId("");
      setIsSubmitting(false);
      setError(null);
    }
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, onClose]);

  if (!open) return null;

  const handleSubmit = async () => {
    const trimmed = channelId.trim();
    if (!trimmed || isSubmitting) return;
    setIsSubmitting(true);
    setError(null);
    try {
      await getChannel().postChannelsSubscribe({
        channel_id: trimmed,
      });
      await mutate(CACHE_KEYS.dashboardSubscriptions);
      onAdded();
      onClose();
    } catch (err) {
      const code = getApiErrorCode(err);
      setError(code ? t(`apiErrors.${code}`, t("apiErrors.fallback")) : t("dashboard.addChannelDialog.error"));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      aria-label={t("dashboard.addChannelDialog.title")}
    >
      <div
        class="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
      />
      <div class="relative bg-white dark:bg-[#2a2721] rounded-2xl shadow-2xl border border-gray-100 dark:border-neutral-800 p-8 max-w-sm w-full">
        <button
          class="absolute top-4 right-4 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors bg-transparent border-none cursor-pointer"
          onClick={onClose}
          aria-label={t("dashboard.addChannelDialog.cancel")}
        >
          <span class="material-symbols-outlined">close</span>
        </button>
        <h2 class="text-2xl font-bold text-charcoal dark:text-white mb-2">
          {t("dashboard.addChannelDialog.title")}
        </h2>
        <p class="text-sm text-text-muted-light dark:text-text-muted-dark mb-4">
          {t("dashboard.addChannelDialog.description")}
        </p>
        <div class="relative">
          <button
            type="button"
            class="absolute inset-y-0 left-0 flex items-center pl-3 pr-1 text-text-muted-light dark:text-text-muted-dark hover:text-primary transition-colors bg-transparent border-none cursor-pointer"
            aria-label={t("dashboard.addChannelDialog.paste")}
            onClick={async () => {
              try {
                const text = await navigator.clipboard.readText();
                if (text) setChannelId(text);
              } catch {}
            }}
          >
            <span class="material-symbols-outlined text-[20px]">
              content_paste
            </span>
          </button>
          <input
            type="text"
            class="w-full pl-10 pr-4 py-3 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white placeholder-taupe focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all shadow-sm"
            placeholder={t("dashboard.addChannelDialog.placeholder")}
            value={channelId}
            onInput={(e) =>
              setChannelId((e.target as HTMLInputElement).value)
            }
          />
        </div>
        {error && (
          <p class="text-sm text-red-500 mt-2" role="alert">
            {error}
          </p>
        )}
        <div class="flex justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 rounded-xl text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors bg-transparent border-none cursor-pointer"
            onClick={onClose}
          >
            {t("dashboard.addChannelDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-xl text-sm font-bold text-white bg-primary hover:bg-primary/90 transition-colors border-none cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={!channelId.trim() || isSubmitting}
            onClick={handleSubmit}
          >
            {isSubmitting
              ? t("dashboard.addChannelDialog.adding")
              : t("dashboard.addChannelDialog.add")}
          </button>
        </div>
      </div>
    </div>
  );
}
