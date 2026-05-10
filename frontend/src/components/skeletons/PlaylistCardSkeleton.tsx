import { Skeleton } from "./Skeleton";

export function PlaylistCardSkeleton() {
  return (
    <div class="flex flex-col bg-card-light dark:bg-card-dark rounded-xl border border-transparent overflow-hidden">
      <Skeleton class="aspect-video w-full" />
      <div class="flex flex-col flex-1 p-5 gap-3">
        <Skeleton class="h-6 w-3/4 rounded" />
        <Skeleton class="h-4 w-full rounded" />
        <Skeleton class="h-4 w-5/6 rounded" />
        <div class="flex items-center gap-3 mt-auto">
          <Skeleton class="h-3 w-12 rounded" />
          <Skeleton class="h-3 w-20 rounded" />
          <Skeleton class="h-3 w-20 rounded" />
        </div>
      </div>
    </div>
  );
}

export function ChannelDetailPlaylistCardSkeleton() {
  return (
    <div class="flex-shrink-0 w-56 bg-card-light dark:bg-card-dark rounded-xl border border-transparent overflow-hidden">
      <Skeleton class="aspect-video w-full" />
      <div class="p-3 flex flex-col gap-2">
        <Skeleton class="h-4 w-11/12 rounded" />
        <Skeleton class="h-4 w-2/3 rounded" />
        <Skeleton class="h-3 w-1/3 rounded mt-1" />
      </div>
    </div>
  );
}
