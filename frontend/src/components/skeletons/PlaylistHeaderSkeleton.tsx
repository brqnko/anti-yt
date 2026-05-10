import { Skeleton } from "./Skeleton";

export function PlaylistHeaderSkeleton() {
  return (
    <div class="bg-card-light dark:bg-card-dark rounded-xl border border-border-light dark:border-border-dark mb-8 p-6">
      <div class="flex flex-col sm:flex-row gap-6 items-start">
        <Skeleton class="w-full sm:w-48 aspect-video flex-shrink-0 rounded-lg" />
        <div class="flex-1 min-w-0 flex flex-col gap-3 w-full">
          <Skeleton class="h-7 md:h-8 w-2/3 rounded" />
          <Skeleton class="h-4 w-full rounded" />
          <Skeleton class="h-4 w-5/6 rounded" />
          <div class="flex flex-wrap items-center gap-x-4 gap-y-1 mt-1">
            <Skeleton class="h-3 w-16 rounded" />
            <Skeleton class="h-3 w-24 rounded" />
            <Skeleton class="h-3 w-24 rounded" />
          </div>
        </div>
        <div class="flex gap-2 flex-shrink-0">
          <Skeleton class="h-9 w-24 rounded-lg" />
          <Skeleton class="h-9 w-20 rounded-lg" />
        </div>
      </div>
    </div>
  );
}
