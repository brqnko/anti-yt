import type { JSX } from "preact";

export function Skeleton({
  class: className = "",
  style,
}: {
  class?: string;
  style?: JSX.CSSProperties;
}) {
  return (
    <div
      class={`animate-pulse bg-gray-200 dark:bg-gray-800 ${className}`}
      style={style}
      aria-hidden="true"
    />
  );
}
