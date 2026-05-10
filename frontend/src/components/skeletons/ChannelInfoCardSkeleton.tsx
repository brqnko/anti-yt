import { Skeleton } from "./Skeleton";

export function ChannelInfoCardSkeleton() {
  return (
    <div class="bg-card-light dark:bg-card-dark rounded-xl border border-border-light dark:border-border-dark mb-8 p-6">
      <div class="flex flex-row gap-4 md:gap-6 items-start md:items-center">
        <Skeleton class="size-16 md:size-28 rounded-full shrink-0" />
        <div class="flex-1 min-w-0 flex flex-col md:flex-row md:items-center gap-3">
          <div class="flex-1 min-w-0 flex flex-col gap-2">
            <Skeleton class="h-7 md:h-9 w-1/2 rounded" />
            <div class="flex flex-wrap gap-x-4 gap-y-1">
              <Skeleton class="h-4 w-32 rounded" />
              <Skeleton class="h-4 w-20 rounded" />
              <Skeleton class="h-4 w-24 rounded" />
            </div>
          </div>
          <Skeleton class="h-14 w-44 rounded-xl flex-shrink-0" />
        </div>
      </div>
      <div class="h-px bg-border-light dark:bg-border-dark my-5" />
      <div class="flex flex-col gap-2">
        <Skeleton class="h-3 w-full rounded" />
        <Skeleton class="h-3 w-full rounded" />
        <Skeleton class="h-3 w-2/3 rounded" />
      </div>
    </div>
  );
}
