import { useState, useEffect, useCallback, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { Icon } from "../../components/Icon";

const GRID_SIZE = 36; // 6x6
const REGROW_INTERVAL_MS = 2000;
const MIN_MOWN_FOR_REGROW = 5;

type TileState = "grass" | "mown";

function GrassMowerGame() {
  const { t } = useTranslation();
  const [tiles, setTiles] = useState<TileState[]>(() =>
    Array(GRID_SIZE).fill("mown"),
  );
  const [score, setScore] = useState(0);
  const regrowRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const audioCtxRef = useRef<AudioContext | null>(null);

  const playMowSound = useCallback(() => {
    if (!audioCtxRef.current) {
      audioCtxRef.current = new AudioContext();
    }
    const ctx = audioCtxRef.current;
    const t = ctx.currentTime;
    const duration = 0.08;

    // White noise buffer for the "シャ" texture
    const bufferSize = Math.ceil(ctx.sampleRate * duration);
    const buffer = ctx.createBuffer(1, bufferSize, ctx.sampleRate);
    const data = buffer.getChannelData(0);
    for (let i = 0; i < bufferSize; i++) {
      data[i] = Math.random() * 2 - 1;
    }
    const noise = ctx.createBufferSource();
    noise.buffer = buffer;

    // Highpass filter to sharpen the sound ("クッ")
    const filter = ctx.createBiquadFilter();
    filter.type = "highpass";
    filter.frequency.setValueAtTime(3000, t);
    filter.frequency.exponentialRampToValueAtTime(800, t + duration);

    // Sharp attack, quick decay
    const gain = ctx.createGain();
    gain.gain.setValueAtTime(0.3, t);
    gain.gain.exponentialRampToValueAtTime(0.001, t + duration);

    noise.connect(filter);
    filter.connect(gain);
    gain.connect(ctx.destination);
    noise.start(t);
    noise.stop(t + duration);
  }, []);

  const playGrowSound = useCallback(() => {
    if (!audioCtxRef.current) {
      audioCtxRef.current = new AudioContext();
    }
    const ctx = audioCtxRef.current;
    const t = ctx.currentTime;

    // Soft rising tone for grass sprouting
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.type = "sine";
    osc.frequency.setValueAtTime(300, t);
    osc.frequency.exponentialRampToValueAtTime(600, t + 0.12);
    gain.gain.setValueAtTime(0.001, t);
    gain.gain.linearRampToValueAtTime(0.1, t + 0.04);
    gain.gain.exponentialRampToValueAtTime(0.001, t + 0.15);
    osc.start(t);
    osc.stop(t + 0.15);
  }, []);

  const resetGame = useCallback(() => {
    setTiles(Array(GRID_SIZE).fill("mown"));
    setScore(0);
  }, []);

  const mowTile = useCallback((index: number) => {
    setTiles((prev) => {
      if (prev[index] !== "grass") return prev;
      const next = [...prev];
      next[index] = "mown";
      return next;
    });
    setScore((s) => s + 1);
    playMowSound();
  }, [playMowSound]);

  // Regrow grass periodically (only when page is visible)
  useEffect(() => {
    const startInterval = () => {
      if (regrowRef.current) return;
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
          playGrowSound();
          return next;
        });
      }, REGROW_INTERVAL_MS);
    };

    const stopInterval = () => {
      if (regrowRef.current) {
        clearInterval(regrowRef.current);
        regrowRef.current = null;
      }
    };

    const handleVisibilityChange = () => {
      if (document.hidden) {
        stopInterval();
      } else if (document.hasFocus()) {
        startInterval();
      }
    };

    const handleFocus = () => {
      if (!document.hidden) startInterval();
    };

    const handleBlur = () => {
      stopInterval();
    };

    if (!document.hidden && document.hasFocus()) startInterval();
    document.addEventListener("visibilitychange", handleVisibilityChange);
    window.addEventListener("focus", handleFocus);
    window.addEventListener("blur", handleBlur);

    return () => {
      stopInterval();
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      window.removeEventListener("focus", handleFocus);
      window.removeEventListener("blur", handleBlur);
    };
  }, [playGrowSound]);

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
      <div class="grid grid-cols-6 gap-2 bg-card-light dark:bg-card-dark p-4 rounded-3xlborder border-border-light dark:border-border-dark w-full aspect-square max-w-[340px]">
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
        <div class="bg-card-light dark:bg-card-dark px-4 py-2 rounded-2xlborder border-border-light dark:border-border-dark flex items-center gap-2">
          <Icon name="grass" class="text-green-500 text-sm" />
          <span class="text-sm font-medium text-charcoal dark:text-white/80">
            {score}
          </span>
        </div>
        <button
          class="bg-card-light dark:bg-card-dark px-4 py-2 rounded-2xlborder border-border-light dark:border-border-dark flex items-center gap-2 hover:bg-border-light/50 dark:hover:bg-border-dark/50 transition-colors cursor-pointer"
          onClick={resetGame}
        >
          <Icon name="refresh" class="text-taupe text-sm" />
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
    <div class="min-h-dvh bg-background-light dark:bg-background-dark flex flex-col items-center justify-center px-6 py-12">
      {/* Header */}
      <div class="text-center mb-10">
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

      {/* Settings link */}
      <a
        href="/screen-time-settings"
        class="mt-8 flex items-center gap-2 text-sm text-taupe dark:text-white/50 hover:text-charcoal dark:hover:text-white/80 transition-colors no-underline"
      >
        {t("screenTimeBlock.goToSettings")}
      </a>
    </div>
  );
}
