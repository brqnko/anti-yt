import { Skeleton } from "./Skeleton";

export function VideoPlayerSkeleton({ hasPlaylist = false }: { hasPlaylist?: boolean }) {
  return (
    <div class="flex-1 overflow-y-auto">
      <div class="max-w-[1536px] mx-auto px-0 sm:px-6 py-0 sm:py-8 pb-8 flex flex-col xl:flex-row xl:items-start gap-8">
        <div class="flex-1 min-w-0">
          <Skeleton class="w-full aspect-video bg-gray-300 dark:bg-gray-900" />

          <div class="mt-8 px-4 sm:px-0 flex flex-col gap-4">
            <Skeleton class="h-7 w-3/4 rounded" />
            <Skeleton class="h-7 w-1/2 rounded" />

            <div class="mt-4 pb-6 border-b border-border-light dark:border-border-dark">
              <div class="flex items-center gap-4">
                <Skeleton class="size-12 rounded-full flex-shrink-0" />
                <div class="flex flex-col gap-2">
                  <Skeleton class="h-5 w-40 rounded" />
                  <Skeleton class="h-4 w-32 rounded" />
                </div>
              </div>
            </div>

            <div class="flex items-center gap-6 mt-3 pb-3 border-b border-border-light dark:border-border-dark">
              {[0, 1, 2, 3, 4, 5].map((i) => (
                <div key={i} class="flex flex-col items-center gap-1">
                  <Skeleton class="size-6 rounded" />
                  <Skeleton class="h-2.5 w-10 rounded" />
                </div>
              ))}
            </div>

            <div class="mt-6">
              <div class="bg-border-light/50 dark:bg-[#332e27]/30 p-6 rounded-xl flex flex-col gap-3">
                <Skeleton class="h-5 w-32 rounded" />
                <Skeleton class="h-4 w-full rounded" />
                <Skeleton class="h-4 w-11/12 rounded" />
                <Skeleton class="h-4 w-3/4 rounded" />
              </div>
            </div>
          </div>
        </div>

        {hasPlaylist && (
          <aside class="w-full xl:w-[420px] shrink-0 flex flex-col gap-8 px-4 sm:px-0">
            <div class="bg-card-light dark:bg-card-dark rounded-2xl border border-border-light dark:border-border-dark flex flex-col overflow-hidden">
              <div class="p-4 border-b border-border-light dark:border-border-dark flex items-center gap-2">
                <Skeleton class="size-5 rounded flex-shrink-0" />
                <div class="flex flex-col gap-2 min-w-0 flex-1">
                  <Skeleton class="h-4 w-2/3 rounded" />
                  <Skeleton class="h-3 w-1/4 rounded" />
                </div>
              </div>
              <div class="flex flex-col">
                {[0, 1, 2, 3, 4].map((i) => (
                  <div key={i} class="flex gap-3 p-3 pr-9">
                    <Skeleton class="h-3 w-5 rounded flex-shrink-0 mt-1" />
                    <Skeleton class="w-20 aspect-video rounded-md flex-shrink-0" />
                    <div class="flex flex-col justify-center min-w-0 flex-1 gap-1.5">
                      <Skeleton class="h-3 w-11/12 rounded" />
                      <Skeleton class="h-3 w-2/3 rounded" />
                      <Skeleton class="h-2.5 w-1/3 rounded mt-0.5" />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </aside>
        )}
      </div>
    </div>
  );
}
