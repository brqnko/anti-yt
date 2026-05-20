import { useState } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { getAuth } from "../../api/generated/auth";
import { getApiErrorCode } from "../../utils/api-error";

export default function Reactivation() {
  const { t } = useTranslation();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleReactivate = async () => {
    setSubmitting(true);
    setError(null);
    try {
      const { postAuthReactivate } = getAuth();
      await postAuthReactivate();
      window.location.href = "/";
    } catch (err) {
      const code = getApiErrorCode(err);
      const msg = code
        ? t(`apiErrors.${code}`, t("apiErrors.fallback"))
        : t("apiErrors.fallback");
      setError(msg);
      setSubmitting(false);
    }
  };

  if (submitting) {
    return null;
  }

  if (error) {
    return (
      <div class="bg-background-light dark:bg-background-dark min-h-dvh flex items-center justify-center px-4">
        <div class="text-center max-w-sm">
          <p class="text-lg font-bold text-charcoal dark:text-white mb-2">
            {t("reactivation.error.title")}
          </p>
          <p class="text-sm text-taupe dark:text-gray-400 mb-6">{error}</p>
          <a
            href="/"
            class="px-6 py-3 bg-primary text-white font-bold rounded-xl inline-block"
          >
            {t("common.returnHome")}
          </a>
        </div>
      </div>
    );
  }

  return (
    <div class="bg-background-light dark:bg-background-dark min-h-dvh flex items-center justify-center px-4">
      <div class="text-center max-w-sm">
        <p class="text-lg font-bold text-charcoal dark:text-white mb-2">
          {t("reactivation.title")}
        </p>
        <p class="text-sm text-taupe dark:text-gray-400 mb-6">
          {t("reactivation.description")}
        </p>
        <button
          onClick={handleReactivate}
          class="w-full px-6 py-3 bg-primary text-white font-bold rounded-xl"
        >
          {t("reactivation.reactivate")}
        </button>
      </div>
    </div>
  );
}
