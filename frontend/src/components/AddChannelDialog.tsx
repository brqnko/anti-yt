import { useState, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { getChannel } from "../api/generated/channel";
import type { PostChannelsSubscribe201 } from "../api/generated/antiYtApi.schemas";
import { getApiErrorCode } from "../utils/api-error";
import { useNotification } from "../contexts/NotificationContext";
import { Dialog } from "./Dialog";
import { Icon } from "./Icon";

interface AddChannelDialogProps {
  open: boolean;
  onClose: () => void;
  onAdded: (channel: PostChannelsSubscribe201) => void;
}

export function AddChannelDialog({
  open,
  onClose,
  onAdded,
}: AddChannelDialogProps) {
  const { t } = useTranslation();
  const { show } = useNotification();
  const [channelId, setChannelId] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!open) {
      setChannelId("");
      setIsSubmitting(false);
    }
  }, [open]);

  if (!open) return null;

  const handleSubmit = async () => {
    const trimmed = channelId.trim();
    if (!trimmed || isSubmitting) return;
    setIsSubmitting(true);
    try {
      const res = await getChannel().postChannelsSubscribe({
        channel_id: trimmed,
      });
      onAdded(res);
      onClose();
    } catch (err) {
      const code = getApiErrorCode(err);
      show({ type: "error", messageKey: code ? `apiErrors.${code}` : "channels.addChannelDialog.error" });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} ariaLabel={t("channels.addChannelDialog.title")} maxWidth="max-w-sm" showCloseButton closeButtonLabel={t("channels.addChannelDialog.cancel")}>
        <h2 class="text-2xl font-bold text-charcoal dark:text-white mb-2">
          {t("channels.addChannelDialog.title")}
        </h2>
        <p class="text-sm text-text-muted-light dark:text-text-muted-dark mb-4">
          {t("channels.addChannelDialog.description")}
        </p>
        <div class="relative">
          <button
            type="button"
            class="absolute inset-y-0 left-0 flex items-center pl-3 pr-1 text-text-muted-light dark:text-text-muted-dark hover:text-primary bg-transparent border-none cursor-pointer"
            aria-label={t("channels.addChannelDialog.paste")}
            onClick={async () => {
              try {
                const text = await navigator.clipboard.readText();
                if (text) setChannelId(text);
              } catch {}
            }}
          >
            <Icon name="content_paste" class="text-[20px]" />
          </button>
          <input
            type="text"
            class="w-full pl-10 pr-4 py-3 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white placeholder-taupe focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all"
            placeholder={t("channels.addChannelDialog.placeholder")}
            value={channelId}
            onInput={(e) =>
              setChannelId((e.target as HTMLInputElement).value)
            }
          />
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 rounded-xl text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 bg-transparent border-none cursor-pointer"
            onClick={onClose}
          >
            {t("channels.addChannelDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-xl text-sm font-bold text-white bg-primary hover:bg-primary/90 border-none cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={!channelId.trim() || isSubmitting}
            onClick={handleSubmit}
          >
            {isSubmitting
              ? t("channels.addChannelDialog.adding")
              : t("channels.addChannelDialog.add")}
          </button>
        </div>
    </Dialog>
  );
}
