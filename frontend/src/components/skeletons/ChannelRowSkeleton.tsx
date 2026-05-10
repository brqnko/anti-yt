import { Skeleton } from "./Skeleton";

export function ChannelRowSkeleton() {
  return (
    <div class="flex items-center gap-4 p-3 rounded-lg bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark">
      <Skeleton class="size-12 rounded-full shrink-0" />
      <div class="flex flex-col gap-2 grow min-w-0">
        <Skeleton class="h-4 w-1/2 rounded" />
        <Skeleton class="h-3 w-2/3 rounded" />
      </div>
      <Skeleton class="size-8 rounded-full shrink-0" />
    </div>
  );
}
