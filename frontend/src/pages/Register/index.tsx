import { useEffect, useState } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { getUser } from "../../api/generated/user";
import { getApiErrorCode } from "../../utils/api-error";

export default function Register() {
  const { t } = useTranslation();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    const register = async () => {
      try {
        const { postUsersMe } = getUser();
        const lang = navigator.language?.startsWith("ja") ? "ja" : "en";
        await postUsersMe({
          display_name: "anonymous",
          language_code: lang,
          screen_time: [{ start_time: "00:00", end_time: "24:00" }],
        });
        if (cancelled) return;
        window.location.href = "/";
      } catch (err) {
        if (cancelled) return;
        const code = getApiErrorCode(err);
        if (code === "user.already_registered") {
          window.location.href = "/";
          return;
        }
        const msg = code
          ? t(`apiErrors.${code}`, t("apiErrors.fallback"))
          : t("apiErrors.fallback");
        setError(msg);
      }
    };

    register();
    return () => { cancelled = true; };
  }, []);

  if (error) {
    return (
      <div class="bg-background-light dark:bg-background-dark min-h-dvh flex items-center justify-center px-4">
        <div class="text-center max-w-sm">
          <p class="text-lg font-bold text-charcoal dark:text-white mb-2">
            {t("register.error.title")}
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
    <div class="bg-background-light dark:bg-background-dark min-h-dvh flex items-center justify-center">
      <div class="animate-spin size-8 border-4 border-primary border-t-transparent rounded-full" />
    </div>
  );
}
