import { Skeleton } from "./Skeleton";

export function VideoCardSkeleton({ layout = "card" }: { layout?: "card" | "row" }) {
  if (layout === "row") {
    return (
      <article class="flex flex-col sm:flex-row gap-3 sm:gap-4">
        <Skeleton class="aspect-video w-full sm:w-48 sm:flex-shrink-0 md:w-60 rounded-xl sm:rounded-lg" />
        <div class="flex flex-col gap-2 sm:gap-3 min-w-0 flex-1 sm:py-1">
          <Skeleton class="h-5 sm:h-6 w-11/12 rounded" />
          <Skeleton class="h-5 sm:h-6 w-2/3 rounded" />
          <div class="flex items-center gap-2 mt-1">
            <Skeleton class="size-5 rounded-full" />
            <Skeleton class="h-3 w-32 rounded" />
            <Skeleton class="h-3 w-16 rounded" />
          </div>
        </div>
      </article>
    );
  }

  return (
    <article class="flex flex-col gap-3">
      <Skeleton class="aspect-video w-full rounded-xl" />
      <div class="flex gap-3 items-start">
        <Skeleton class="size-9 rounded-full flex-shrink-0" />
        <div class="flex flex-col min-w-0 flex-1 gap-2">
          <Skeleton class="h-4 w-11/12 rounded" />
          <Skeleton class="h-4 w-2/3 rounded" />
          <Skeleton class="h-3 w-1/2 rounded mt-1" />
        </div>
      </div>
    </article>
  );
}
