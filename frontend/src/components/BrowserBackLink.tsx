import type { ComponentChildren, JSX } from "preact";

type BrowserBackLinkProps = Omit<
  JSX.HTMLAttributes<HTMLAnchorElement>,
  "href"
> & {
  children: ComponentChildren;
  fallbackHref: string;
};

export function BrowserBackLink({
  children,
  fallbackHref,
  onClick,
  ...props
}: BrowserBackLinkProps) {
  const handleClick = (event: JSX.TargetedMouseEvent<HTMLAnchorElement>) => {
    onClick?.(event);
    if (
      event.defaultPrevented ||
      event.button !== 0 ||
      event.metaKey ||
      event.ctrlKey ||
      event.shiftKey ||
      event.altKey
    ) {
      return;
    }

    event.preventDefault();
    if (window.history.length > 1) {
      window.history.back();
      return;
    }
    window.location.href = fallbackHref;
  };

  return (
    <a href={fallbackHref} onClick={handleClick} {...props}>
      {children}
    </a>
  );
}
