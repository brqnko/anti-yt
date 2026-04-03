import type { VNode } from "preact";

const URL_REGEX = /(https?:\/\/[^\s<>"']+)/g;
const URL_TEST = /^https?:\/\//;
const TIMESTAMP_REGEX = /(?:(\d{1,2}):)?(\d{1,2}):(\d{2})/;
const TOKEN_REGEX = new RegExp(
  `(https?:\\/\\/[^\\s<>"']+)|(?:(\\d{1,2}):)?(\\d{1,2}):(\\d{2})`,
  "g",
);

function parseTimestamp(
  hours: string | undefined,
  minutes: string,
  seconds: string,
): number {
  return (
    (hours ? parseInt(hours, 10) * 3600 : 0) +
    parseInt(minutes, 10) * 60 +
    parseInt(seconds, 10)
  );
}

export function Linkify({
  text,
  onTimestamp,
}: {
  text: string;
  onTimestamp?: (seconds: number) => void;
}): VNode {
  if (!onTimestamp) {
    // Original URL-only behaviour
    const parts = text.split(URL_REGEX);
    return (
      <>
        {parts.map((part, i) =>
          URL_TEST.test(part) ? (
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

  // URL + timestamp handling
  const nodes: (string | VNode)[] = [];
  let lastIndex = 0;

  TOKEN_REGEX.lastIndex = 0;
  let match: RegExpExecArray | null;
  while ((match = TOKEN_REGEX.exec(text)) !== null) {
    // Push text before this match
    if (match.index > lastIndex) {
      nodes.push(text.slice(lastIndex, match.index));
    }

    if (match[1]) {
      // URL match
      nodes.push(
        <a
          key={match.index}
          href={match[1]}
          target="_blank"
          rel="noopener noreferrer"
          class="text-primary hover:underline break-all"
        >
          {match[1]}
        </a>,
      );
    } else {
      // Timestamp match: groups [2]=hours, [3]=minutes, [4]=seconds
      const seconds = parseTimestamp(match[2], match[3], match[4]);
      const label = match[0];
      nodes.push(
        <button
          key={match.index}
          type="button"
          class="text-primary hover:underline bg-transparent border-none cursor-pointer p-0 font-inherit text-inherit"
          onClick={() => onTimestamp(seconds)}
        >
          {label}
        </button>,
      );
    }

    lastIndex = match.index + match[0].length;
  }

  // Push remaining text
  if (lastIndex < text.length) {
    nodes.push(text.slice(lastIndex, text.length));
  }

  return <>{nodes}</>;
}
