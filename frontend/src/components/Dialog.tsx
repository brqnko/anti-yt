import type { ComponentChildren } from "preact";
import { useEscapeKey } from "../hooks/useEscapeKey";

interface DialogProps {
  open: boolean;
  onClose: () => void;
  ariaLabel: string;
  /** Max-width class, e.g. "max-w-md", "max-w-sm". Defaults to "max-w-md". */
  maxWidth?: string;
  /** Extra classes for the panel div. */
  panelClass?: string;
  children: ComponentChildren;
}

export function Dialog({
  open,
  onClose,
  ariaLabel,
  maxWidth = "max-w-md",
  panelClass,
  children,
}: DialogProps) {
  useEscapeKey(open, onClose);

  if (!open) return null;

  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      aria-label={ariaLabel}
    >
      <div class="absolute inset-0 bg-black/60" onClick={onClose} />
      <div
        class={`relative bg-white dark:bg-[#2a2721] rounded-2xl ring-1 ring-black/10 dark:ring-white/10 border border-gray-100 dark:border-neutral-800 p-8 ${maxWidth} w-full ${panelClass ?? ""}`}
      >
        {children}
      </div>
    </div>
  );
}
