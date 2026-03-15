import type { VNode } from "preact";

const URL_REGEX = /(https?:\/\/[^\s<>\"']+)/g;

export function Linkify({ text }: { text: string }): VNode {
  const parts = text.split(URL_REGEX);
  return (
    <>
      {parts.map((part, i) =>
        URL_REGEX.test(part) ? (
          <a
            key={i}
            href={part}
            target="_blank"
            rel="noopener noreferrer"
            class="text-primary hover:underline break-all"
          >
            {part}
          </a>
        ) : (
          part
        ),
      )}
    </>
  );
}
