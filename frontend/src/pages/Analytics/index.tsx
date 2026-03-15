import { useTranslation, Trans } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";

function AnalyticsContent() {
  const { t } = useTranslation();
  useTitle(t("analytics.pageTitle"));

  return (
    <DashboardLayout>
      <div class="flex-grow w-full flex justify-center py-8 px-4 md:px-8 lg:px-20">
        <div class="flex flex-col w-full max-w-[1024px] gap-8">
          {/* Page Heading */}
          <div class="flex flex-col gap-2">
            <h1 class="text-4xl font-black leading-tight tracking-[-0.033em] text-charcoal dark:text-white">
              {t("analytics.title")}
            </h1>
          </div>

          {/* Stats Cards */}
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Time Wasted */}
            <div class="flex flex-col gap-3 rounded-xl p-6 border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark shadow-sm">
              <p class="text-text-muted-light dark:text-text-muted-dark text-sm font-medium uppercase tracking-wider">
                {t("analytics.timeWasted")}
              </p>
              <div class="flex items-baseline gap-2">
                <p class="text-3xl font-bold text-charcoal dark:text-white">42h 15m</p>
              </div>
              <div class="flex items-center gap-1 text-green-600 dark:text-green-400 text-sm font-medium">
                <span class="material-symbols-outlined text-lg">trending_down</span>
                <p>{t("analytics.lessVsLastMonth", { percent: 12 })}</p>
              </div>
            </div>

            {/* Daily Average */}
            <div class="flex flex-col gap-3 rounded-xl p-6 border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark shadow-sm">
              <p class="text-text-muted-light dark:text-text-muted-dark text-sm font-medium uppercase tracking-wider">
                {t("analytics.dailyAverage")}
              </p>
              <div class="flex items-baseline gap-2">
                <p class="text-3xl font-bold text-charcoal dark:text-white">1h 24m</p>
              </div>
              <div class="flex items-center gap-1 text-green-600 dark:text-green-400 text-sm font-medium">
                <span class="material-symbols-outlined text-lg">trending_down</span>
                <p>{t("analytics.lessVsLastMonth", { percent: 5 })}</p>
              </div>
            </div>

            {/* Goal Progress */}
            <div class="flex flex-col gap-3 rounded-xl p-6 border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark shadow-sm">
              <p class="text-text-muted-light dark:text-text-muted-dark text-sm font-medium uppercase tracking-wider">
                {t("analytics.goalProgress")}
              </p>
              <div class="flex items-baseline gap-2">
                <p class="text-3xl font-bold text-charcoal dark:text-white">85%</p>
              </div>
              <div class="flex items-center gap-1 text-primary text-sm font-medium">
                <span class="material-symbols-outlined text-lg">check_circle</span>
                <p>{t("analytics.onTrackToMeetGoal")}</p>
              </div>
            </div>
          </div>

          {/* AI Assistant Message */}
          <div class="rounded-xl p-6 bg-primary/10 dark:bg-[#2d2820] border border-primary/20 dark:border-primary/10 flex items-start gap-5 relative overflow-hidden">
            <div class="absolute -right-10 -top-10 text-primary/5 dark:text-primary/5 pointer-events-none">
              <span class="material-symbols-outlined text-[180px]">psychology</span>
            </div>
            <div class="relative z-10 bg-primary/20 dark:bg-primary/30 rounded-full p-3 shrink-0">
              <span class="material-symbols-outlined text-primary text-2xl">psychology</span>
            </div>
            <div class="flex flex-col gap-2 relative z-10">
              <p class="text-text-muted-light dark:text-text-muted-dark text-xs font-bold uppercase tracking-widest">
                {t("analytics.focusAI")}
              </p>
              <p class="text-lg md:text-xl font-medium leading-relaxed text-charcoal dark:text-white">
                <Trans
                  i18nKey="analytics.insightMessage"
                  values={{ hours: 42, less: 5 }}
                  components={{ strong: <strong class="text-primary font-bold" /> }}
                />
              </p>
            </div>
          </div>

          {/* Main Chart Section */}
          <div class="flex flex-col rounded-xl border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark shadow-sm overflow-hidden">
            <div class="p-6 border-b border-border-light dark:border-border-dark flex justify-between items-center flex-wrap gap-4">
              <h3 class="text-lg font-bold text-charcoal dark:text-white">
                {t("analytics.monthlyUsageTrends")}
              </h3>
            </div>
            <div class="p-6">
              {/* Chart Container */}
              <div class="relative h-64 w-full flex items-end justify-between gap-2 md:gap-4 pt-8">
                {/* Limit Line (Dashed) */}
                <div class="absolute top-[30%] left-0 w-full border-t-2 border-dashed border-gray-300 dark:border-gray-600 z-0">
                  <span class="absolute -top-6 right-0 text-xs font-bold text-gray-400 uppercase tracking-wider">
                    {t("analytics.dailyLimit", { hours: 2 })}
                  </span>
                </div>

                {/* Week 1 */}
                <div class="relative z-10 flex flex-col items-center gap-2 h-full justify-end flex-1 group">
                  <div
                    class="w-full max-w-[60px] bg-primary/80 dark:bg-primary/60 rounded-t-md hover:bg-primary transition-all duration-300 relative group-hover:shadow-lg"
                    style="height: 45%"
                  >
                    <div class="absolute -top-8 left-1/2 -translate-x-1/2 bg-gray-900 text-white text-xs py-1 px-2 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                      {t("analytics.weekHours", { week: 1, hours: 12 })}
                    </div>
                  </div>
                  <p class="text-xs font-bold text-text-muted-light dark:text-text-muted-dark uppercase tracking-widest">
                    {t("analytics.weekLabel", { week: 1 })}
                  </p>
                </div>

                {/* Week 2 */}
                <div class="relative z-10 flex flex-col items-center gap-2 h-full justify-end flex-1 group">
                  <div
                    class="w-full max-w-[60px] bg-primary/80 dark:bg-primary/60 rounded-t-md hover:bg-primary transition-all duration-300 relative group-hover:shadow-lg"
                    style="height: 65%"
                  >
                    <div class="absolute -top-8 left-1/2 -translate-x-1/2 bg-gray-900 text-white text-xs py-1 px-2 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                      {t("analytics.weekHours", { week: 2, hours: 18 })}
                    </div>
                  </div>
                  <p class="text-xs font-bold text-text-muted-light dark:text-text-muted-dark uppercase tracking-widest">
                    {t("analytics.weekLabel", { week: 2 })}
                  </p>
                </div>

                {/* Week 3 */}
                <div class="relative z-10 flex flex-col items-center gap-2 h-full justify-end flex-1 group">
                  <div
                    class="w-full max-w-[60px] bg-primary/80 dark:bg-primary/60 rounded-t-md hover:bg-primary transition-all duration-300 relative group-hover:shadow-lg"
                    style="height: 30%"
                  >
                    <div class="absolute -top-8 left-1/2 -translate-x-1/2 bg-gray-900 text-white text-xs py-1 px-2 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                      {t("analytics.weekHours", { week: 3, hours: 8 })}
                    </div>
                  </div>
                  <p class="text-xs font-bold text-text-muted-light dark:text-text-muted-dark uppercase tracking-widest">
                    {t("analytics.weekLabel", { week: 3 })}
                  </p>
                </div>

                {/* Week 4 (Current) */}
                <div class="relative z-10 flex flex-col items-center gap-2 h-full justify-end flex-1 group">
                  <div
                    class="w-full max-w-[60px] bg-primary/40 dark:bg-primary/30 rounded-t-md border-2 border-dashed border-primary hover:bg-primary/50 transition-all duration-300 relative group-hover:shadow-lg"
                    style="height: 15%"
                  >
                    <div class="absolute -top-8 left-1/2 -translate-x-1/2 bg-gray-900 text-white text-xs py-1 px-2 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                      {t("analytics.weekHoursCurrent", { week: 4, hours: 4 })}
                    </div>
                  </div>
                  <p class="text-xs font-bold text-text-muted-light dark:text-text-muted-dark uppercase tracking-widest">
                    {t("analytics.weekLabel", { week: 4 })}
                  </p>
                </div>
              </div>
            </div>
          </div>

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
