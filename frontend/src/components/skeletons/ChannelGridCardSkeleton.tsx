import { Skeleton } from "./Skeleton";

export function ChannelGridCardSkeleton() {
  return (
    <div class="bg-card-light dark:bg-card-dark p-6 rounded-xl border border-border-light dark:border-border-dark flex flex-col gap-4 min-h-[240px]">
      <div class="flex items-center gap-3">
        <Skeleton class="size-12 rounded-full" />
        <div class="flex flex-col gap-2">
          <Skeleton class="h-4 w-40 rounded" />
          <Skeleton class="h-4 w-20 rounded-full" />
        </div>
      </div>
      <div class="flex flex-col gap-2">
        <Skeleton class="h-3 w-full rounded" />
        <Skeleton class="h-3 w-full rounded" />
        <Skeleton class="h-3 w-11/12 rounded" />
        <Skeleton class="h-3 w-4/5 rounded" />
        <Skeleton class="h-3 w-3/5 rounded" />
      </div>
    </div>
  );
}
