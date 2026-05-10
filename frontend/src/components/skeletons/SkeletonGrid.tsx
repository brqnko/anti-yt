import type { ComponentChildren } from "preact";

export function SkeletonRepeat({
  count,
  render,
}: {
  count: number;
  render: (index: number) => ComponentChildren;
}) {
  const items = [];
  for (let i = 0; i < count; i++) {
    items.push(render(i));
  }
  return <>{items}</>;
}
