import { useState, useEffect } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { AxiosError } from "axios";
import { useTitle } from "../../hooks/useTitle";
import { useAuth } from "../../contexts/AuthContext";
import { getUser } from "../../api/generated/user";
import type { ProblemDetailError } from "../../api/generated/antiYtApi.schemas";
import Step1 from "./Step1";
import Step2 from "./Step2";
import type { TimeRange } from "../../types/time-range";
import { formatTime } from "../../utils/format";

export default function Register() {
  const { t, i18n } = useTranslation();
  const { route } = useLocation();
  const { refreshAuth } = useAuth();
  useTitle(t("register.pageTitle"));
  const [step, setStep] = useState(1);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<(ProblemDetailError & { apiDetail?: string }) | null>(null);

  // Step 1 data
  const [displayName, setDisplayName] = useState("");
  const [languageCode, setLanguageCode] = useState("en");

  // Step 2 data
  const [isLimited, setIsLimited] = useState(true);
  const [hours, setHours] = useState(1);
  const [minutes, setMinutes] = useState(30);
  const [timeRanges, setTimeRanges] = useState<TimeRange[]>([]);

  useEffect(() => {
    setLanguageCode(i18n.language.startsWith("ja") ? "ja" : "en");
  }, []);

  const handleRegister = async () => {
    if (submitting) return;
    setSubmitting(true);
    try {
      const { postUsersMe } = getUser();
      await postUsersMe({
        display_name: displayName.trim(),
        language_code: languageCode,
        screen_time: timeRanges.map((r) => ({
          id: r.id,
          start_time: formatTime(r.startMinutes),
          end_time: formatTime(r.endMinutes),
        })),
        daily_screen_seconds: isLimited ? hours * 3600 + minutes * 60 : undefined,
      });
      await refreshAuth();
      route("/dashboard");
    } catch (err) {
      const apiDetail =
        err instanceof AxiosError && err.response?.data?.detail
          ? String(err.response.data.detail)
          : undefined;
      setError({ title: t("register.error.title"), detail: t("register.error.description"), apiDetail });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div class="bg-background-light dark:bg-background-dark min-h-screen flex flex-col font-display text-charcoal dark:text-gray-100 transition-colors duration-200">
      <main class="flex-grow w-full flex items-center justify-center px-4 sm:px-6 py-24">
        <div class="w-full max-w-2xl">
          {step === 1 ? (
            <Step1
              displayName={displayName}
              setDisplayName={setDisplayName}
              languageCode={languageCode}
              setLanguageCode={setLanguageCode}
              onNext={() => setStep(2)}
            />
          ) : (
            <Step2
              isLimited={isLimited}
              setIsLimited={setIsLimited}
              hours={hours}
              setHours={setHours}
              minutes={minutes}
              setMinutes={setMinutes}
              timeRanges={timeRanges}
              setTimeRanges={setTimeRanges}
              submitting={submitting}
              onBack={() => setStep(1)}
              onNext={handleRegister}
            />
          )}
        </div>
      </main>

      {/* Error dialog */}
      {error && (
        <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={() => setError(null)} />
          <div class="relative bg-white dark:bg-[#2a2721] rounded-2xl shadow-2xl border border-gray-100 dark:border-neutral-800 p-8 max-w-sm w-full text-center">
            <button
              type="button"
              class="absolute top-4 right-4 text-taupe hover:text-charcoal dark:hover:text-white transition-colors cursor-pointer"
              onClick={() => setError(null)}
            >
              <span class="material-symbols-outlined">close</span>
            </button>
            <h3 class="text-xl font-black text-charcoal dark:text-white mb-2">
              {error.title}
            </h3>
            <p class="text-sm text-taupe dark:text-gray-400 mb-2">
              {error.detail}
            </p>
            {error.apiDetail && (
              <p class="text-xs text-taupe/70 dark:text-gray-500 mb-6 font-mono">
                {error.apiDetail}
              </p>
            )}
            {!error.apiDetail && <div class="mb-4" />}
            <button
              type="button"
              class="w-full px-6 py-3 bg-primary hover:bg-[#b8a37e] text-charcoal font-bold rounded-xl transition-all cursor-pointer"
              onClick={() => setError(null)}
            >
              {t("register.error.close")}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
