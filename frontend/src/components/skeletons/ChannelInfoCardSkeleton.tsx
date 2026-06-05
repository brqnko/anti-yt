import { Skeleton } from "./Skeleton";

export function ChannelInfoCardSkeleton() {
  return (
    <section class="mb-8">
      <div class="flex flex-col sm:flex-row gap-5 md:gap-6 items-start">
        <Skeleton class="size-20 md:size-24 rounded-full shrink-0" />
        <div class="flex-1 min-w-0 w-full flex flex-col gap-3">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="flex-1 min-w-0 flex flex-col gap-2">
              <Skeleton class="h-7 md:h-9 w-1/2 rounded" />
              <div class="flex flex-wrap gap-x-4 gap-y-1">
                <Skeleton class="h-4 w-32 rounded" />
                <Skeleton class="h-4 w-24 rounded" />
              </div>
            </div>
            <div class="flex items-center gap-3 flex-shrink-0">
              <Skeleton class="h-4 w-28 rounded" />
              <Skeleton class="h-7 w-14 rounded-full" />
            </div>
          </div>
          <div class="flex flex-col gap-2 mt-1">
            <Skeleton class="h-3 w-full rounded" />
            <Skeleton class="h-3 w-full rounded" />
            <Skeleton class="h-3 w-2/3 rounded" />
          </div>
        </div>
      </div>
    </section>
  );
}
