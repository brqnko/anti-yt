import { Skeleton } from "./Skeleton";

const BAR_HEIGHTS = [40, 65, 50, 85, 70, 55, 90];

export function AnalyticsSkeleton() {
  return (
    <>
      <div class="rounded-xl p-6 bg-primary/10 dark:bg-[#2d2820] border border-primary/20 dark:border-primary/10 flex flex-col gap-3">
        <Skeleton class="h-5 md:h-6 w-full rounded" />
        <Skeleton class="h-5 md:h-6 w-11/12 rounded" />
        <Skeleton class="h-5 md:h-6 w-2/3 rounded" />
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[0, 1, 2].map((i) => (
          <div
            key={i}
            class="flex flex-col gap-3 rounded-xl p-6 border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark"
          >
            <Skeleton class="h-4 w-1/2 rounded" />
            <Skeleton class="h-8 w-2/3 rounded" />
          </div>
        ))}
      </div>

      <div class="flex flex-col rounded-xl border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark">
        <div class="p-6 border-b border-border-light dark:border-border-dark">
          <Skeleton class="h-6 w-48 rounded" />
        </div>
        <div class="p-6">
          <div class="relative h-64 w-full flex items-end justify-between gap-2 md:gap-4 pt-8">
            {BAR_HEIGHTS.map((h, i) => (
              <div
                key={i}
                class="relative z-10 flex flex-col items-center gap-2 h-full justify-end flex-1"
              >
                <Skeleton
                  class="w-full max-w-[60px] rounded-t-md"
                  style={{ height: `${h}%` }}
                />
                <Skeleton class="h-3 w-8 rounded" />
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  );
}
