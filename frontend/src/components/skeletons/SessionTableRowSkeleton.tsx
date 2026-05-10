import { Skeleton } from "./Skeleton";

export function SessionTableRowSkeleton() {
  return (
    <tr class="border-b border-border-light dark:border-border-dark last:border-0">
      <td class="px-4 py-3">
        <Skeleton class="h-4 w-24 rounded" />
      </td>
      <td class="px-4 py-3">
        <Skeleton class="h-4 w-32 rounded" />
      </td>
      <td class="px-4 py-3">
        <Skeleton class="h-4 w-48 rounded" />
      </td>
      <td class="px-4 py-3">
        <Skeleton class="h-4 w-28 rounded" />
      </td>
      <td class="px-4 py-3">
        <Skeleton class="h-4 w-28 rounded" />
      </td>
      <td class="px-4 py-3 w-10">
        <Skeleton class="h-6 w-6 rounded" />
      </td>
    </tr>
  );
}
