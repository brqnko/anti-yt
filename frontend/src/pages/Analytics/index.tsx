import { useCallback, useEffect, useMemo, useRef, useState } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { getHistory } from "../../api/generated/history";
import { getUser } from "../../api/generated/user";
import { toDateStr, getLastNDays, isoToDateStr, formatDateLabel } from "../../utils/format";
import type { GetStatisticsWeekly200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";

function AnalyticsContent() {
  const { t } = useTranslation();
  useTitle(t("analytics.pageTitle"));

  const [items, setItems] = useState<GetStatisticsWeekly200ItemsItem[]>([]);
  const [aiSummary, setAiSummary] = useState<string | undefined>();
  const [dailyLimitSeconds, setDailyLimitSeconds] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(false);

  const load = useCallback(async () => {
    setIsLoading(true);
    setError(false);
    try {
      const startDay = getLastNDays(7)[0];
      const [weeklyData, userData] = await Promise.all([
        getHistory().getStatisticsWeekly({ target_week: toDateStr(startDay) }),
        getUser().getUsersMeStatus(),
      ]);
      setItems(weeklyData.items);
      setAiSummary(weeklyData.ai_summary);
      setDailyLimitSeconds(userData.daily_screen_seconds ?? null);
    } catch {
      setError(true);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const dailyLimitHours = dailyLimitSeconds != null ? dailyLimitSeconds / 3600 : null;

  const dailyBars = useMemo(() => {
    const days = getLastNDays(7);
    const todayStr = toDateStr(new Date());
    return days.map((date) => {
      const dayStr = toDateStr(date);
      const item = items.find((it) => isoToDateStr(it.target_day) === dayStr);
      const seconds = item?.video_watch_seconds ?? 0;
      const hours = seconds / 3600;
      const videos = item?.video_watch_count ?? 0;
      const isToday = dayStr === todayStr;
      const label = formatDateLabel(date);
      return { hours, seconds, videos, label, isToday };
    });
  }, [items, isLoading]);

  const totalSeconds = useMemo(
    () => items.reduce((s, d) => s + d.video_watch_seconds, 0),
    [items],
  );
  const totalVideos = useMemo(
    () => items.reduce((s, d) => s + d.video_watch_count, 0),
    [items],
  );
  const daysWithData = useMemo(
    () => items.filter((d) => d.video_watch_seconds > 0).length,
    [items],
  );
  const dailyAverageSeconds = daysWithData > 0 ? totalSeconds / daysWithData : 0;

  const formatHM = (secs: number) => {
    const h = Math.floor(secs / 3600);
    const m = Math.floor((secs % 3600) / 60);
    if (h > 0 && m > 0) return `${h}h ${m}m`;
    if (h > 0) return `${h}h`;
    return `${m}m`;
  };

  const maxHours = useMemo(() => {
    const dataMax = Math.max(...dailyBars.map((d) => d.hours), 0);
    const limit = dailyLimitHours ?? 0;
    return Math.max(Math.ceil(Math.max(dataMax, limit) * 1.2), 1);
  }, [dailyBars, dailyLimitHours]);

  const [activeBar, setActiveBar] = useState<number | null>(null);
  const chartRef = useRef<HTMLDivElement>(null);

  const dismissTooltip = useCallback((e: MouseEvent) => {
    if (chartRef.current && !chartRef.current.contains(e.target as Node)) {
      setActiveBar(null);
    }
  }, []);

  useEffect(() => {
    document.addEventListener("click", dismissTooltip);
    return () => document.removeEventListener("click", dismissTooltip);
  }, [dismissTooltip]);


  return (
    <DashboardLayout>
      <div class="flex-grow w-full flex justify-center py-8 px-4 md:px-8 lg:px-20">
        <div class="flex flex-col w-full max-w-[1024px] gap-8">
          <div class="flex flex-col gap-2">
            <h1 class="text-4xl font-black leading-tight tracking-[-0.033em] text-charcoal dark:text-white">
              {t("analytics.title")}
            </h1>
          </div>

          {isLoading ? null : error ? (
            <div class="flex flex-col items-center justify-center py-20 text-text-muted-light dark:text-text-muted-dark">
              <Icon name="error_outline" class="text-5xl mb-4" />
              <p class="text-lg font-medium">{t("analytics.loadError")}</p>
              <button onClick={load} class="mt-4 text-sm text-primary hover:underline">
                {t("analytics.retry")}
              </button>
            </div>
          ) : (
            <>
          {aiSummary && (
            <div class="rounded-xl p-6 bg-primary/10 dark:bg-[#2d2820] border border-primary/20 dark:border-primary/10 relative overflow-hidden">
              <p class="text-lg md:text-xl font-medium leading-relaxed text-charcoal dark:text-white">
                {aiSummary}
              </p>
            </div>
          )}

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="flex flex-col gap-3 rounded-xl p-6 border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark">
              <p class="text-text-muted-light dark:text-text-muted-dark text-sm font-medium uppercase tracking-wider">
                {t("analytics.timeWasted")}
              </p>
              <p class="text-3xl font-bold text-charcoal dark:text-white">
                {formatHM(totalSeconds)}
              </p>
            </div>

            <div class="flex flex-col gap-3 rounded-xl p-6 border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark">
              <p class="text-text-muted-light dark:text-text-muted-dark text-sm font-medium uppercase tracking-wider">
                {t("analytics.dailyAverage")}
              </p>
              <p class="text-3xl font-bold text-charcoal dark:text-white">
                {formatHM(dailyAverageSeconds)}
              </p>
            </div>

            <div class="flex flex-col gap-3 rounded-xl p-6 border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark">
              <p class="text-text-muted-light dark:text-text-muted-dark text-sm font-medium uppercase tracking-wider">
                {t("analytics.totalVideos")}
              </p>
              <p class="text-3xl font-bold text-charcoal dark:text-white">
                {totalVideos}{t("analytics.totalVideosUnit")}
              </p>
            </div>
          </div>

          <div ref={chartRef} class="flex flex-col rounded-xl border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark">
            <div class="p-6 border-b border-border-light dark:border-border-dark flex justify-between items-center flex-wrap gap-4">
              <h3 class="text-lg font-bold text-charcoal dark:text-white">
                {t("analytics.weeklyUsageTrends")}
              </h3>
            </div>
            <div class="p-6">
              <div class="relative h-64 w-full flex items-end justify-between gap-2 md:gap-4 pt-8">
                {dailyLimitHours != null && (
                  <div
                    class="absolute left-0 w-full border-t-2 border-dashed border-gray-300 dark:border-gray-600 z-0"
                    style={{ top: `${(1 - dailyLimitHours / maxHours) * 100}%` }}
                  >
                    <span class="absolute -top-6 right-0 text-xs font-bold text-gray-400 uppercase tracking-wider">
                      {t("analytics.dailyLimit", { hours: Math.round(dailyLimitHours * 10) / 10 })}
                    </span>
                  </div>
                )}

                {dailyBars.map((bar, i) => {
                  const pct = maxHours > 0 ? Math.min((bar.hours / maxHours) * 100, 100) : 0;
                  const isActive = activeBar === i;
                  const toggle = () => setActiveBar(isActive ? null : i);
                  return (
                    <div
                      key={bar.label}
                      class="relative z-10 flex flex-col items-center gap-2 h-full justify-end flex-1 group cursor-pointer"
                      role="button"
                      tabIndex={0}
                      aria-label={`${bar.label}: ${formatHM(bar.seconds)}, ${t("analytics.tooltipVideos", { count: bar.videos })}`}
                      onClick={toggle}
                      onKeyDown={(e: KeyboardEvent) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault();
                          toggle();
                        }
                      }}
                    >
                      <div
                        class={`w-full max-w-[60px] rounded-t-md relative ${
                          bar.isToday
                            ? "bg-primary/40 dark:bg-primary/30 border-2 border-dashed border-primary hover:bg-primary/50"
                            : "bg-primary/80 dark:bg-primary/60 hover:bg-primary"
                        }`}
                        style={{
                          height: `${Math.max(pct, 3)}%`,
                        }}
                      >
                        <div
                          class={`absolute bottom-full left-1/2 -translate-x-1/2 -translate-y-2 bg-gray-900 text-white text-xs py-1.5 px-2.5 rounded whitespace-nowrap flex flex-col items-center gap-0.5 pointer-events-none z-20 ${
                            isActive ? "opacity-100" : "opacity-0 group-hover:opacity-100"
                          }`}
                          role="tooltip"
                        >
                          <span>{t("analytics.tooltipVideos", { count: bar.videos })}</span>
                          <span>{formatHM(bar.seconds)}</span>
                        </div>
                      </div>
                      <p class={`text-xs font-bold tracking-wider ${
                        bar.isToday
                          ? "text-primary"
                          : "text-text-muted-light dark:text-text-muted-dark"
                      }`}>
                        {bar.label}
                      </p>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
            </>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
}

export default function Analytics() {
  return (
    <ProtectedRoute>
      <AnalyticsContent />
    </ProtectedRoute>
  );
}
