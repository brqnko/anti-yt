import { useState, useEffect, useCallback, useRef, useMemo } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";

const GRID_SIZE = 36; // 6x6
const REGROW_INTERVAL_MS = 5000;
const MIN_MOWN_FOR_REGROW = 5;

type TileState = "grass" | "mown";

function GrassMowerGame() {
  const { t } = useTranslation();
  const [tiles, setTiles] = useState<TileState[]>(() =>
    Array(GRID_SIZE).fill("grass"),
  );
  const mownCount = useMemo(() => tiles.filter((t) => t === "mown").length, [tiles]);
  const regrowRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const resetGame = useCallback(() => {
    setTiles(Array(GRID_SIZE).fill("grass"));
  }, []);

  const mowTile = useCallback((index: number) => {
    setTiles((prev) => {
      if (prev[index] !== "grass") return prev;
      const next = [...prev];
      next[index] = "mown";
      return next;
    });
  }, []);

  // Regrow grass periodically
  useEffect(() => {
    regrowRef.current = setInterval(() => {
      setTiles((prev) => {
        const mownIndices = prev
          .map((s, i) => (s === "mown" ? i : -1))
          .filter((i) => i >= 0);
        if (mownIndices.length < MIN_MOWN_FOR_REGROW) return prev;
        const randomIdx =
          mownIndices[Math.floor(Math.random() * mownIndices.length)];
        const next = [...prev];
        next[randomIdx] = "grass";
        return next;
      });
    }, REGROW_INTERVAL_MS);

    return () => {
      if (regrowRef.current) clearInterval(regrowRef.current);
    };
  }, []);

  return (
    <div class="flex flex-col items-center">
      <div class="text-center mb-6">
        <h2 class="text-lg font-medium text-charcoal dark:text-white/90">
          {t("screenTimeBlock.mowTheLawn")}
        </h2>
        <p class="text-sm text-taupe dark:text-white/50">
          {t("screenTimeBlock.tapTheGrass")}
        </p>
      </div>

      {/* Grass Grid */}
      <div class="grid grid-cols-6 gap-2 bg-card-light dark:bg-card-dark p-4 rounded-3xl shadow-sm border border-border-light dark:border-border-dark w-full aspect-square max-w-[340px]">
        {tiles.map((state, i) => (
          <button
            key={i}
            class={`rounded-lg aspect-square border-none cursor-pointer transition-all duration-300 active:scale-95 ${
              state === "grass"
                ? "bg-green-400 hover:bg-green-500"
                : "bg-[#e5d3b3] dark:bg-[#4a3f30] cursor-default"
            }`}
            onClick={() => mowTile(i)}
            disabled={state === "mown"}
            aria-label={
              state === "grass"
                ? t("screenTimeBlock.mowTile")
                : t("screenTimeBlock.mownTile")
            }
          />
        ))}
      </div>

      {/* Stats row */}
      <div class="mt-6 flex gap-4">
        <div class="bg-card-light dark:bg-card-dark px-4 py-2 rounded-2xl shadow-sm border border-border-light dark:border-border-dark flex items-center gap-2">
          <span class="material-symbols-outlined text-green-500 text-sm">
            grass
          </span>
          <span class="text-sm font-medium text-charcoal dark:text-white/80">
            {mownCount}
          </span>
        </div>
        <button
          class="bg-card-light dark:bg-card-dark px-4 py-2 rounded-2xl shadow-sm border border-border-light dark:border-border-dark flex items-center gap-2 hover:bg-border-light/50 dark:hover:bg-border-dark/50 transition-colors cursor-pointer"
          onClick={resetGame}
        >
          <span class="material-symbols-outlined text-taupe text-sm">
            refresh
          </span>
          <span class="text-sm font-medium text-charcoal dark:text-white/80">
            {t("screenTimeBlock.reset")}
          </span>
        </button>
      </div>
    </div>
  );
}

export default function ScreenTimeBlock({
  reason,
}: {
  reason: "limit_exceeded" | "outside_time_range";
}) {
  const { t } = useTranslation();
  useTitle(t("screenTimeBlock.pageTitle"));

  return (
    <div class="min-h-screen bg-background-light dark:bg-background-dark flex flex-col items-center justify-center px-6 py-12">
      {/* Header */}
      <div class="text-center mb-10">
        <div class="inline-flex items-center justify-center size-16 rounded-full bg-primary/10 mb-4">
          <span class="material-symbols-outlined text-primary text-3xl">
            timer_off
          </span>
        </div>
        <h1 class="text-2xl font-bold text-charcoal dark:text-white tracking-tight">
          {t("screenTimeBlock.title")}
        </h1>
        <p class="text-sm text-taupe dark:text-white/50 mt-2 max-w-md mx-auto">
          {reason === "limit_exceeded"
            ? t("screenTimeBlock.limitExceeded")
            : t("screenTimeBlock.outsideTimeRange")}
        </p>
      </div>

      {/* Game */}
      <div class="w-full max-w-md">
        <GrassMowerGame />
      </div>

      {/* Zen label */}
      <p class="text-xs text-taupe dark:text-white/30 uppercase tracking-widest font-medium mt-10">
        {t("screenTimeBlock.zenMode")}
      </p>
    </div>
  );
}
