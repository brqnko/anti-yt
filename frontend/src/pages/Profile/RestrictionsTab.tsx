import { useState, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { getUser } from "../../api/generated/user";
import { getApiErrorCode } from "../../utils/api-error";
import { TimeRangeSlider } from "../../components/TimeRangeSlider";
import { formatTime, parseTimeToMinutes } from "../../utils/format";
import type { TimeRange } from "../../types/time-range";
import { Icon } from "../../components/Icon";

export function RestrictionsTab() {
  const { t } = useTranslation();

  // Time ranges from user's screen_time
  const [timeRanges, setTimeRanges] = useState<TimeRange[]>([]);

  // Daily cap
  const [hours, setHours] = useState(1);
  const [minutes, setMinutes] = useState(0);
  const [isUnlimited, setIsUnlimited] = useState(false);

  useEffect(() => {
    const { getUsersMeStatus } = getUser();
    getUsersMeStatus().then((user) => {
      setTimeRanges(
        (user.screen_time ?? []).map((slot, i) => ({
          id: String(i),
          startMinutes: parseTimeToMinutes(slot.start_time),
          endMinutes: parseTimeToMinutes(slot.end_time),
        })),
      );
      const ds = user.daily_screen_seconds;
      if (ds != null && ds > 0) {
        setHours(Math.floor(ds / 3600));
        setMinutes(Math.floor((ds % 3600) / 60));
        setIsUnlimited(false);
      } else if (ds == null) {
        setIsUnlimited(true);
      }
    }).catch(() => {});
  }, []);

  // Saving state
  const [isSaving, setIsSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveFading, setSaveFading] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const addRange = () => {
    setTimeRanges([
      ...timeRanges,
      { id: crypto.randomUUID(), startMinutes: 540, endMinutes: 660 },
    ]);
  };

  const updateRange = (id: string, updated: TimeRange) => {
    setTimeRanges(timeRanges.map((r) => (r.id === id ? updated : r)));
  };

  const deleteRange = (id: string) => {
    setTimeRanges(timeRanges.filter((r) => r.id !== id));
  };

  const handleSave = async () => {
    setIsSaving(true);
    setSaveSuccess(false);
    setSaveError(null);
    try {
      const { patchUsersMeStatus } = getUser();
      const updated = await patchUsersMeStatus({
        screen_time: timeRanges.map((r) => ({
          start_time: formatTime(r.startMinutes),
          end_time: formatTime(r.endMinutes),
        })),
        daily_screen_seconds: isUnlimited ? 86400 : hours * 3600 + minutes * 60,
      });
      setTimeRanges(
        (updated.screen_time ?? []).map((slot, i) => ({
          id: String(i),
          startMinutes: parseTimeToMinutes(slot.start_time),
          endMinutes: parseTimeToMinutes(slot.end_time),
        })),
      );
      const ds = updated.daily_screen_seconds;
      if (ds != null && ds > 0) {
        setHours(Math.floor(ds / 3600));
        setMinutes(Math.floor((ds % 3600) / 60));
        setIsUnlimited(false);
      } else if (ds == null) {
        setIsUnlimited(true);
      }
      setSaveSuccess(true);
      setSaveFading(false);
      setTimeout(() => setSaveFading(true), 2500);
      setTimeout(() => { setSaveSuccess(false); setSaveFading(false); }, 3000);
    } catch (err) {
      const code = getApiErrorCode(err);
      setSaveError(code ? t(`apiErrors.${code}`, t("apiErrors.fallback")) : t("restrictions.saveError"));
    } finally {
      setIsSaving(false);
    }
  };

  const clampHours = (v: number) =>
    setHours(Math.max(0, Math.min(23, isNaN(v) ? 0 : v)));
  const clampMinutes = (v: number) =>
    setMinutes(Math.max(0, Math.min(59, isNaN(v) ? 0 : v)));

  const isTimeInvalid = !isUnlimited && hours === 0 && minutes === 0;

  return (
    <div class="flex flex-col gap-6 min-w-0 overflow-hidden">
          {/* Time Constraints */}
          <div class="flex flex-col rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark overflow-hidden">
            <div class="p-6 border-b border-border-light dark:border-border-dark">
              <h2 class="text-xl font-bold">
                {t("restrictions.timeConstraints")}
              </h2>
            </div>
            <div class="p-6 flex flex-col gap-8">
              {/* Permitted Hours */}
              <div class="flex flex-col gap-6">
                <div class="flex justify-between items-center flex-wrap gap-2">
                  <label class="text-base font-semibold">
                    {t("restrictions.permittedHours")}
                  </label>
                  <span class="text-xs font-medium text-text-muted-light dark:text-text-muted-dark">
                    {t("restrictions.permittedHoursDesc")}
                  </span>
                </div>
                <div class="flex flex-col gap-4">
                  {timeRanges.map((range) => (
                    <TimeRangeSlider
                      key={range.id}
                      range={range}
                      onChange={(updated) => updateRange(range.id, updated)}
                      onDelete={() => deleteRange(range.id)}
                    />
                  ))}
                </div>
                <button
                  type="button"
                  class="flex items-center justify-center gap-2 w-full py-3 border border-dashed border-border-light dark:border-border-dark rounded-lg text-text-muted-light dark:text-text-muted-dark hover:border-primary hover:text-primary hover:bg-primary/5 group cursor-pointer bg-transparent"
                  onClick={addRange}
                >
                  <Icon name="add" />
                  <span class="text-sm font-bold">
                    {t("restrictions.addTimeRange")}
                  </span>
                </button>
              </div>

              <hr class="border-border-light dark:border-border-dark" />

              {/* Daily Cap Limit */}
              <div class="flex flex-col gap-4">
                <label class="text-base font-semibold">
                  {t("restrictions.dailyCapLimit")}
                </label>
                <div class="flex flex-wrap gap-4 items-center">
                  {/* Unlimited toggle */}
                  <button
                    type="button"
                    role="switch"
                    aria-checked={!isUnlimited}
                    onClick={() => setIsUnlimited(!isUnlimited)}
                    class={`relative inline-flex h-7 w-12 shrink-0 items-center rounded-full transition-colors cursor-pointer border-none ${
                      !isUnlimited ? "bg-primary" : "bg-gray-300 dark:bg-gray-600"
                    }`}
                  >
                    <span
                      class={`inline-block size-5 rounded-full bg-white transition-transform ${
                        !isUnlimited ? "translate-x-6" : "translate-x-1"
                      }`}
                    />
                  </button>
                  <span class="text-sm font-medium text-text-muted-light dark:text-text-muted-dark">
                    {t("restrictions.enableLimit")}
                  </span>

                  <div class="w-px h-6 bg-border-light dark:bg-border-dark" />
                  <div class={`flex flex-wrap gap-4 items-center transition-opacity ${isUnlimited ? "opacity-40 pointer-events-none" : ""}`}>
                    <div class="relative">
                      <input
                        type="number"
                        class="w-24 pl-4 pr-8 py-3 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary outline-none text-center font-bold text-lg"
                        min={0}
                        max={23}
                        value={hours}
                        disabled={isUnlimited}
                        onInput={(e) =>
                          clampHours(
                            parseInt((e.target as HTMLInputElement).value),
                          )
                        }
                      />
                      <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-text-muted-light dark:text-text-muted-dark font-medium">
                        {t("restrictions.hr")}
                      </span>
                    </div>
                    <span class="text-text-muted-light dark:text-text-muted-dark font-bold">
                      :
                    </span>
                    <div class="relative">
                      <input
                        type="number"
                        class="w-24 pl-4 pr-8 py-3 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary outline-none text-center font-bold text-lg"
                        min={0}
                        max={59}
                        value={minutes}
                        disabled={isUnlimited}
                        onInput={(e) =>
                          clampMinutes(
                            parseInt((e.target as HTMLInputElement).value),
                          )
                        }
                      />
                      <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-text-muted-light dark:text-text-muted-dark font-medium">
                        {t("restrictions.min")}
                      </span>
                    </div>
                  </div>
                </div>
                {!isUnlimited && isTimeInvalid && (
                  <p class="text-sm text-red-500">
                    {t("restrictions.zeroTimeError")}
                  </p>
                )}
                <div class="flex items-center gap-3">
                  <button
                    class="px-6 py-3 bg-primary hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-lg flex items-center gap-2 cursor-pointer border-none"
                    disabled={isSaving || isTimeInvalid}
                    onClick={handleSave}
                  >
                    {isSaving ? t("restrictions.saving") : t("restrictions.save")}
                  </button>
                  {saveSuccess && (
                    <span class={`text-sm text-green-600 dark:text-green-400 font-medium transition-opacity duration-500 ${saveFading ? "opacity-0" : "opacity-100"}`}>
                      {t("restrictions.saved")}
                    </span>
                  )}
                  {saveError && (
                    <span class="text-sm text-red-500 font-medium flex items-center gap-1">
                      <Icon name="error" class="text-[18px]" />
                      {saveError}
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>
    </div>
  );
}
