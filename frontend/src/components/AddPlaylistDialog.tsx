import { useState, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { mutate } from "swr";
import { getPlaylist } from "../api/generated/playlist";
import type { PostPlaylists201 } from "../api/generated/antiYtApi.schemas";
import { CACHE_KEYS } from "../api/cache-keys";
import { getApiErrorCode } from "../utils/api-error";
import { Dialog } from "./Dialog";
import { Icon } from "./Icon";

interface AddPlaylistDialogProps {
  open: boolean;
  onClose: () => void;
  onAdded: (playlist: PostPlaylists201) => void;
}

export function AddPlaylistDialog({
  open,
  onClose,
  onAdded,
}: AddPlaylistDialogProps) {
  const { t } = useTranslation();
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [importUrl, setImportUrl] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) {
      setTitle("");
      setDescription("");
      setImportUrl("");
      setIsSubmitting(false);
      setError(null);
    }
  }, [open]);

  if (!open) return null;

  const handleSubmit = async () => {
    const trimmedTitle = title.trim();
    if (!trimmedTitle || isSubmitting) return;
    setIsSubmitting(true);
    setError(null);
    try {
      const res = await getPlaylist().postPlaylists({
        playlist_title: trimmedTitle,
        playlist_description: description.trim(),
        playlist_type: "normal",
        playlist_visibility: "private",
        ...(importUrl.trim() ? { base_playlist_url: importUrl.trim() } : {}),
      });
      await mutate(CACHE_KEYS.dashboardPlaylists);
      onAdded(res);
      onClose();
    } catch (err) {
      const code = getApiErrorCode(err);
      setError(code ? t(`apiErrors.${code}`, t("apiErrors.fallback")) : t("dashboard.addPlaylistDialog.error"));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} ariaLabel={t("dashboard.addPlaylistDialog.title")} showCloseButton closeButtonLabel={t("dashboard.addPlaylistDialog.cancel")}>
        <h2 class="text-2xl font-bold text-charcoal dark:text-white mb-5">
          {t("dashboard.addPlaylistDialog.title")}
        </h2>

        <div class="flex flex-col gap-4">
          {/* Title */}
          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium text-charcoal dark:text-white">
              {t("dashboard.addPlaylistDialog.titleLabel")}
            </label>
            <input
              type="text"
              class="w-full px-4 py-3 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white placeholder-taupe focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all"
              placeholder={`${t("dashboard.addPlaylistDialog.titlePlaceholder")} (${t("dashboard.addPlaylistDialog.required")})`}
              value={title}
              onInput={(e) => setTitle((e.target as HTMLInputElement).value)}
            />
          </div>

          {/* Description */}
          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium text-charcoal dark:text-white">
              {t("dashboard.addPlaylistDialog.descriptionLabel")}
            </label>
            <textarea
              class="w-full px-4 py-3 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white placeholder-taupe focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all resize-none"
              rows={3}
              placeholder={t("dashboard.addPlaylistDialog.descriptionPlaceholder")}
              value={description}
              onInput={(e) => setDescription((e.target as HTMLTextAreaElement).value)}
            />
          </div>

          {/* Import from YouTube playlist URL (optional) */}
          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium text-charcoal dark:text-white">
              {t("dashboard.addPlaylistDialog.importLabel")}
              <span class="text-text-muted-light dark:text-text-muted-dark font-normal ml-1">
                ({t("dashboard.addPlaylistDialog.optional")})
              </span>
            </label>
            <div class="relative">
              <button
                type="button"
                class="absolute inset-y-0 left-0 flex items-center pl-3 pr-1 text-text-muted-light dark:text-text-muted-dark hover:text-primary bg-transparent border-none cursor-pointer"
                onClick={async () => {
                  try {
                    const text = await navigator.clipboard.readText();
                    if (text) setImportUrl(text);
                  } catch {}
                }}
              >
                <Icon name="content_paste" class="text-[20px]" />
              </button>
              <input
                type="text"
                class="w-full pl-10 pr-4 py-3 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white placeholder-taupe focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all"
                placeholder={t("dashboard.addPlaylistDialog.importPlaceholder")}
                value={importUrl}
                onInput={(e) => setImportUrl((e.target as HTMLInputElement).value)}
              />
            </div>
          </div>
        </div>

        {error && (
          <p class="text-sm text-red-500 mt-3" role="alert">
            {error}
          </p>
        )}

        <div class="flex justify-end gap-3 mt-6">
          <button
            class="px-4 py-2 rounded-xl text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 bg-transparent border-none cursor-pointer"
            onClick={onClose}
          >
            {t("dashboard.addPlaylistDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-xl text-sm font-bold text-white bg-primary hover:bg-primary/90 border-none cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={!title.trim() || isSubmitting}
            onClick={handleSubmit}
          >
            {isSubmitting
              ? t("dashboard.addPlaylistDialog.adding")
              : t("dashboard.addPlaylistDialog.add")}
          </button>
        </div>
    </Dialog>
  );
}
