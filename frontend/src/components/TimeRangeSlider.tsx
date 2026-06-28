import { useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { formatTime } from "../utils/format";
import type { TimeRange } from "../types/time-range";
import { Icon } from "./Icon";

interface TimeRangeSliderProps {
  range: TimeRange;
  onChange: (r: TimeRange) => void;
  onDelete: () => void;
  variant?: "register" | "settings";
}

export function TimeRangeSlider({
  range,
  onChange,
  onDelete,
  variant = "settings",
}: TimeRangeSliderProps) {
  const { t } = useTranslation();
  const trackRef = useRef<HTMLDivElement>(null);
  const draggingRef = useRef<"start" | "end" | null>(null);

  const totalMinutes = 1440;
  const startPct = (range.startMinutes / totalMinutes) * 100;
  const endPct = (range.endMinutes / totalMinutes) * 100;

  const calcMinutes = (clientX: number): number => {
    if (!trackRef.current) return 0;
    const rect = trackRef.current.getBoundingClientRect();
    const x = Math.max(0, Math.min(clientX - rect.left, rect.width));
    const raw = (x / rect.width) * totalMinutes;
    return Math.max(0, Math.min(Math.round(raw / 15) * 15, totalMinutes));
  };

  const handlePointerDown = (thumb: "start" | "end") => (e: PointerEvent) => {
    e.preventDefault();
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    draggingRef.current = thumb;
  };

  const handlePointerMove = (e: PointerEvent) => {
    const thumb = draggingRef.current;
    if (!thumb) return;
    const mins = calcMinutes(e.clientX);
    if (thumb === "start") {
      onChange({ ...range, startMinutes: Math.min(mins, range.endMinutes - 15) });
    } else {
      onChange({ ...range, endMinutes: Math.max(mins, range.startMinutes + 15) });
    }
  };

  const handlePointerUp = (e: PointerEvent) => {
    (e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId);
    draggingRef.current = null;
  };

  const thumbs = (
    <>
      <div
        class={`absolute top-1/2 -mt-2 -ml-2 size-4 bg-white ${variant === "register" ? "dark:bg-[#2a2721]" : "dark:bg-card-dark"} border-2 border-primary rounded-full cursor-grab z-10 active:scale-95 transition-transform touch-none before:content-[''] before:absolute before:-inset-3`}
        style={{ left: `${startPct}%` }}
        onPointerDown={handlePointerDown("start")}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
      />
      <div
        class={`absolute top-1/2 -mt-2 -ml-2 size-4 bg-white ${variant === "register" ? "dark:bg-[#2a2721]" : "dark:bg-card-dark"} border-2 border-primary rounded-full cursor-grab z-10 active:scale-95 transition-transform touch-none before:content-[''] before:absolute before:-inset-3`}
        style={{ left: `${endPct}%` }}
        onPointerDown={handlePointerDown("end")}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
      />
    </>
  );

  const axisLabels = (
    <div class={`flex justify-between text-[10px] font-medium ${variant === "register" ? "w-full mt-2 text-taupe px-0.5 select-none pointer-events-none" : "text-text-muted-light dark:text-text-muted-dark mb-2"}`}>
      <span>00:00</span>
      <span class="hidden sm:inline">06:00</span>
      <span>12:00</span>
      <span class="hidden sm:inline">18:00</span>
      <span>24:00</span>
    </div>
  );

  if (variant === "register") {
    return (
      <div class="group relative flex items-center gap-4">
        <div class="flex-grow pt-8 pb-4 relative select-none">
          <div
            class="absolute top-1 z-20 transition-all duration-200"
            style={{ left: `${startPct}%`, transform: "translateX(-50%)" }}
          >
            <span class="bg-white dark:bg-[#2a2721] text-primary border border-gray-200 dark:border-neutral-700 text-[10px] font-bold px-1.5 py-0.5 rounded">
              {formatTime(range.startMinutes)}
            </span>
          </div>
          <div
            class="absolute top-1 z-20 transition-all duration-200"
            style={{ left: `${endPct}%`, transform: "translateX(-50%)" }}
          >
            <span class="bg-white dark:bg-[#2a2721] text-primary border border-gray-200 dark:border-neutral-700 text-[10px] font-bold px-1.5 py-0.5 rounded">
              {formatTime(range.endMinutes)}
            </span>
          </div>

          <div
            ref={trackRef}
            class="h-3 w-full bg-gray-100 dark:bg-neutral-800 rounded-full relative border border-gray-200 dark:border-neutral-700"
          >
            <div class="absolute inset-0 w-full flex justify-between px-0.5 items-center opacity-30 pointer-events-none">
              <div class="h-1.5 w-px bg-current text-taupe" />
              <div class="h-1.5 w-px bg-current text-taupe" />
              <div class="h-1.5 w-px bg-current text-taupe" />
              <div class="h-1.5 w-px bg-current text-taupe" />
              <div class="h-1.5 w-px bg-current text-taupe" />
            </div>

            <div
              class="absolute top-0 bottom-0 bg-primary/30 rounded-full pointer-events-none"
              style={{ left: `${startPct}%`, width: `${endPct - startPct}%` }}
            />
            <div
              class="absolute top-[3px] bottom-[3px] bg-primary rounded-full pointer-events-none"
              style={{ left: `${startPct}%`, width: `${endPct - startPct}%` }}
            />

            {thumbs}
          </div>

          {axisLabels}
        </div>

        <button
          type="button"
          class="mt-4 size-10 flex items-center justify-center text-taupe hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-xl border border-transparent hover:border-red-100 dark:hover:border-red-900/30 cursor-pointer"
          title={t("restrictions.removeRange")}
          onClick={onDelete}
        >
          <Icon name="delete" class="text-xl" />
        </button>
      </div>
    );
  }

  return (
    <div class="flex flex-col gap-4 p-4 rounded-xl bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="px-2.5 py-1 bg-primary/10 text-primary text-xs font-bold rounded border border-primary/20">
            {formatTime(range.startMinutes)} - {formatTime(range.endMinutes)}
          </div>
        </div>
        <button
          type="button"
          aria-label={t("restrictions.removeRange")}
          class="size-8 flex items-center justify-center text-text-muted-light dark:text-text-muted-dark hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg cursor-pointer bg-transparent border-none"
          onClick={onDelete}
        >
          <Icon name="delete" class="text-[18px]" />
        </button>
      </div>
      <div class="relative pt-2 px-1">
        {axisLabels}
        <div
          ref={trackRef}
          class="h-1.5 w-full bg-border-light dark:bg-border-dark rounded-full relative"
        >
          <div
            class="absolute h-full bg-primary rounded-full pointer-events-none"
            style={{ left: `${startPct}%`, width: `${endPct - startPct}%` }}
          />
          {thumbs}
        </div>
      </div>
    </div>
  );
}
