import { useTranslation } from "react-i18next";
import { TimeRangeSlider } from "../../components/TimeRangeSlider";
import { hasOverlap, type TimeRange } from "../../types/time-range";

interface Step2Props {
  isLimited: boolean;
  setIsLimited: (v: boolean) => void;
  hours: number;
  setHours: (v: number) => void;
  minutes: number;
  setMinutes: (v: number) => void;
  timeRanges: TimeRange[];
  setTimeRanges: (v: TimeRange[]) => void;
  submitting: boolean;
  onBack: () => void;
  onNext: () => void;
}

export default function Step2({
  isLimited,
  setIsLimited,
  hours,
  setHours,
  minutes,
  setMinutes,
  timeRanges,
  setTimeRanges,
  submitting,
  onBack,
  onNext,
}: Step2Props) {
  const { t } = useTranslation();

  const overlapping = hasOverlap(timeRanges);
  const isTimeInvalid = isLimited && hours === 0 && minutes === 0;
  const canSubmit = !submitting && !overlapping && !isTimeInvalid;

  const updateRange = (id: string, updated: TimeRange) => {
    setTimeRanges(timeRanges.map((r) => (r.id === id ? updated : r)));
  };

  const deleteRange = (id: string) => {
    setTimeRanges(timeRanges.filter((r) => r.id !== id));
  };

  const addRange = () => {
    setTimeRanges([
      ...timeRanges,
      { id: crypto.randomUUID(), startMinutes: 1080, endMinutes: 1260 },
    ]);
  };

  const clampHours = (v: number) => setHours(Math.max(0, Math.min(23, isNaN(v) ? 0 : v)));
  const clampMinutes = (v: number) => setMinutes(Math.max(0, Math.min(59, isNaN(v) ? 0 : v)));

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    if (!canSubmit) return;
    onNext();
  };

  return (
    <>
      {/* Step header */}
      <div class="mb-8 text-center">
        <span class="inline-block py-1 px-3 rounded-full bg-primary/10 text-primary text-xs font-bold uppercase tracking-wider mb-3">
          {t("register.step", { current: 2, total: 2 })}
        </span>
        <h2 class="text-3xl font-black text-charcoal dark:text-white mb-2">
          {t("register.restrictions.title")}
        </h2>
        <p class="text-taupe dark:text-gray-400">
          {t("register.restrictions.subtitle")}
        </p>
      </div>

      {/* Card */}
      <div class="bg-white dark:bg-[#2a2721] rounded-2xl shadow-xl border border-gray-100 dark:border-neutral-800 p-6 md:p-8 relative overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-primary to-transparent opacity-50" />

        <form class="space-y-0" onSubmit={handleSubmit}>
          {/* Daily Watch Limit */}
          <div class="mb-8 pb-8 border-b border-gray-100 dark:border-neutral-800">
            <div class="flex items-center justify-between mb-6">
              <div class="flex items-center gap-4">
                <div class="size-10 rounded-lg bg-orange-50 dark:bg-orange-900/20 text-orange-600 flex items-center justify-center shadow-sm">
                  <span class="material-symbols-outlined">timer</span>
                </div>
                <div>
                  <h3 class="font-bold text-lg text-charcoal dark:text-white leading-tight">
                    {t("register.restrictions.dailyWatchLimit")}
                  </h3>
                  <p class="text-xs text-taupe dark:text-gray-400">
                    {t("register.restrictions.dailyWatchLimitDesc")}
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-3 bg-background-light dark:bg-neutral-800/50 p-1.5 rounded-full border border-gray-100 dark:border-neutral-800">
                <span class="text-[10px] font-bold text-taupe uppercase tracking-wider pl-2">
                  {t("register.restrictions.unlimited")}
                </span>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    class="sr-only peer"
                    checked={isLimited}
                    onChange={(e) => setIsLimited((e.target as HTMLInputElement).checked)}
                  />
                  <div class="w-11 h-6 bg-gray-300 peer-focus:outline-none rounded-full peer dark:bg-neutral-600 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary shadow-sm" />
                </label>
                <span class="text-[10px] font-bold text-primary uppercase tracking-wider pr-2">
                  {t("register.restrictions.limited")}
                </span>
              </div>
            </div>

            {isLimited && (
              <div class="flex flex-col sm:flex-row gap-4 items-center">
                <div class="w-full relative group">
                  <label class="absolute -top-2.5 left-3 bg-white dark:bg-[#2a2721] px-1 text-[10px] uppercase tracking-widest font-bold text-primary z-10">
                    {t("register.restrictions.hours")}
                  </label>
                  <input
                    type="number"
                    class="w-full bg-background-light dark:bg-background-dark border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white text-xl font-bold rounded-xl px-4 py-3 focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-all shadow-inner"
                    min={0}
                    max={23}
                    value={hours}
                    onInput={(e) => clampHours(parseInt((e.target as HTMLInputElement).value))}
                  />
                </div>
                <div class="hidden sm:block text-2xl text-taupe font-light px-2">:</div>
                <div class="w-full relative group">
                  <label class="absolute -top-2.5 left-3 bg-white dark:bg-[#2a2721] px-1 text-[10px] uppercase tracking-widest font-bold text-primary z-10">
                    {t("register.restrictions.minutes")}
                  </label>
                  <input
                    type="number"
                    class="w-full bg-background-light dark:bg-background-dark border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white text-xl font-bold rounded-xl px-4 py-3 focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-all shadow-inner"
                    min={0}
                    max={59}
                    value={minutes}
                    onInput={(e) => clampMinutes(parseInt((e.target as HTMLInputElement).value))}
                  />
                </div>
              </div>
            )}
          </div>

          {/* Scheduled Access */}
          <div class="mb-8">
            <div class="flex items-center justify-between mb-6">
              <div class="flex items-center gap-4">
                <div class="size-10 rounded-lg bg-blue-50 dark:bg-blue-900/20 text-blue-600 flex items-center justify-center shadow-sm">
                  <span class="material-symbols-outlined">calendar_clock</span>
                </div>
                <div>
                  <h3 class="font-bold text-lg text-charcoal dark:text-white leading-tight">
                    {t("register.restrictions.scheduledAccess")}
                  </h3>
                  <p class="text-xs text-taupe dark:text-gray-400">
                    {t("register.restrictions.scheduledAccessDesc")}
                  </p>
                </div>
              </div>
            </div>

            <div class="space-y-6">
              {timeRanges.map((range) => (
                <TimeRangeSlider
                  key={range.id}
                  range={range}
                  onChange={(updated) => updateRange(range.id, updated)}
                  onDelete={() => deleteRange(range.id)}
                  variant="register"
                />
              ))}
            </div>

            <button
              type="button"
              class="mt-4 w-full sm:w-auto flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl border border-dashed border-gray-300 dark:border-neutral-700 text-sm font-bold text-taupe hover:text-primary hover:border-primary hover:bg-primary/5 transition-all cursor-pointer"
              onClick={addRange}
            >
              <span class="material-symbols-outlined text-lg">add_circle</span>
              {t("register.restrictions.addTimeRange")}
            </button>
          </div>

          {/* Validation errors */}
          {(overlapping || isTimeInvalid) && (
            <div class="mt-4 p-3 rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-sm text-red-600 dark:text-red-400 space-y-1">
              {isTimeInvalid && <p>{t("register.restrictions.zeroTimeError")}</p>}
              {overlapping && <p>{t("register.restrictions.overlapError")}</p>}
            </div>
          )}

          {/* Navigation */}
          <div class="mt-8 space-y-4">
            <div class="flex items-center justify-between gap-4">
              <button
                type="button"
                class="px-6 py-3 text-sm font-bold text-taupe hover:text-charcoal dark:hover:text-white transition-colors hover:underline underline-offset-4 cursor-pointer"
                onClick={onBack}
              >
                {t("register.restrictions.back")}
              </button>
              <button
                type="submit"
                disabled={!canSubmit}
                class="bg-primary hover:bg-[#b8a37e] text-charcoal px-8 py-3 rounded-xl font-bold transition-all flex items-center gap-2 active:scale-95 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <span>{t("register.submit")}</span>
                <span class="material-symbols-outlined text-lg">check</span>
              </button>
            </div>
            <p class="text-xs text-taupe text-center">
              {t("register.agreementPrefix")}
              <a href="/terms" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">{t("register.termsLink")}</a>
              {t("register.agreementAnd")}
              <a href="/privacy" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">{t("register.privacyLink")}</a>
              {t("register.agreementSuffix")}
            </p>
          </div>
        </form>
      </div>
    </>
  );
}
