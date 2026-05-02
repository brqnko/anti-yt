import { useState, useRef, useEffect } from "preact/hooks";
import { createPortal } from "preact/compat";
import { Icon } from "./Icon";

export interface DropdownOption {
  value: string;
  label: string;
  icon?: string;
}

interface DropdownProps {
  value: string;
  options: DropdownOption[];
  onChange: (value: string) => void;
  ariaLabel?: string;
  leadingIcon?: string;
  className?: string;
}

export function Dropdown({
  value,
  options,
  onChange,
  ariaLabel,
  leadingIcon,
  className,
}: DropdownProps) {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState({ top: 0, left: 0, width: 0 });
  const triggerRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  const updatePos = () => {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    setPos({
      top: rect.bottom + window.scrollY + 8,
      left: rect.left + window.scrollX,
      width: rect.width,
    });
  };

  useEffect(() => {
    if (!open) return;
    updatePos();
    const handleClick = (e: MouseEvent) => {
      const target = e.target as Node;
      if (
        triggerRef.current &&
        !triggerRef.current.contains(target) &&
        menuRef.current &&
        !menuRef.current.contains(target)
      ) {
        setOpen(false);
      }
    };
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    const handleReposition = () => updatePos();
    document.addEventListener("mousedown", handleClick);
    document.addEventListener("keydown", handleKeyDown);
    window.addEventListener("scroll", handleReposition, true);
    window.addEventListener("resize", handleReposition);
    return () => {
      document.removeEventListener("mousedown", handleClick);
      document.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("scroll", handleReposition, true);
      window.removeEventListener("resize", handleReposition);
    };
  }, [open]);

  const selected = options.find((o) => o.value === value);
  const startIcon = leadingIcon ?? selected?.icon;

  return (
    <div class={`relative ${className ?? ""}`}>
      <button
        ref={triggerRef}
        type="button"
        aria-label={ariaLabel}
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
        class="w-full flex items-center gap-2 pl-10 pr-8 py-2.5 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none font-medium text-left cursor-pointer relative"
      >
        {startIcon && (
          <span class="absolute inset-y-0 left-3 flex items-center text-text-muted-light dark:text-text-muted-dark pointer-events-none">
            <Icon name={startIcon} class="text-[20px]" />
          </span>
        )}
        <span class="truncate">{selected?.label ?? ""}</span>
        <span class="absolute inset-y-0 right-3 flex items-center text-text-muted-light dark:text-text-muted-dark pointer-events-none">
          <Icon name="expand_more" class="text-[20px]" />
        </span>
      </button>
      {open &&
        typeof document !== "undefined" &&
        createPortal(
          <div
            ref={menuRef}
            role="listbox"
            style={{
              position: "absolute",
              top: pos.top,
              left: pos.left,
              width: pos.width,
            }}
            class="z-[200] py-2 bg-white dark:bg-[#1a1a1a] rounded-xl ring-1 ring-black/5 dark:ring-white/5 border border-slate-200 dark:border-white/10"
          >
            {options.map((opt) => (
              <button
                role="option"
                type="button"
                key={opt.value}
                aria-selected={opt.value === value}
                class={`w-full flex items-center gap-2 px-4 py-2 text-left text-sm font-medium cursor-pointer transition-colors ${
                  opt.value === value
                    ? "text-primary"
                    : "text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-white/5"
                }`}
                onClick={() => {
                  onChange(opt.value);
                  setOpen(false);
                }}
              >
                <Icon
                  name="check"
                  class={`text-base ${opt.value === value ? "opacity-100" : "opacity-0"}`}
                />
                {opt.icon && (
                  <Icon name={opt.icon} class="text-[18px] shrink-0" />
                )}
                <span class="truncate">{opt.label}</span>
              </button>
            ))}
          </div>,
          document.body,
        )}
    </div>
  );
}
