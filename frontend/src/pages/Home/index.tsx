import { useTranslation } from "react-i18next";
import { GoogleIcon } from "./GoogleIcon";

function DashboardPreview({ t }: { t: (key: string) => string }) {
  return (
    <div class="animate-fade-in-up relative w-[90%] max-w-4xl bg-white dark:bg-[#151515] rounded-2xl shadow-2xl border border-slate-200 dark:border-white/10 overflow-hidden">
      <div class="p-8 bg-slate-50/50 dark:bg-[#0f0f0f]">
        <div class="grid grid-cols-2 gap-6">
          {/* Allowance card */}
          <div class="col-span-2 bg-white dark:bg-[#1a1a1a] p-6 rounded-xl border border-slate-200 dark:border-white/5 shadow-sm flex items-center justify-between">
            <div>
              <h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-1">
                {t("home.dailyAllowance")}
              </h3>
              <div class="text-3xl font-extrabold flex items-baseline gap-1">
                00:45:00{" "}
                <span class="text-sm font-normal text-slate-400">/ 1h 00m</span>
              </div>
            </div>
            <div class="relative w-16 h-16 flex items-center justify-center">
              <svg class="-rotate-90 w-full h-full" viewBox="0 0 36 36">
                <path
                  class="text-slate-200 dark:text-white/10"
                  d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="4"
                />
                <path
                  class="text-primary"
                  d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                  fill="none"
                  stroke="currentColor"
                  stroke-dasharray="75, 100"
                  stroke-width="4"
                />
              </svg>
              <span class="material-symbols-outlined absolute text-primary text-xl">
                timer
              </span>
            </div>
          </div>

          {/* Chart card */}
          <div class="bg-white dark:bg-[#1a1a1a] p-6 rounded-xl border border-slate-200 dark:border-white/5 shadow-sm">
            <h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-4">
              {t("home.watchTime")}
            </h3>
            <div class="flex items-end justify-between h-32 gap-2">
              <div
                class="flex-1 bg-primary/30 rounded-t-sm"
                style="height:30%"
              />
              <div
                class="flex-1 bg-primary/50 rounded-t-sm"
                style="height:50%"
              />
              <div
                class="flex-1 bg-primary/60 rounded-t-sm"
                style="height:40%"
              />
              <div
                class="flex-1 bg-primary/40 rounded-t-sm"
                style="height:60%"
              />
              <div
                class="flex-1 bg-primary rounded-t-sm relative"
                style="height:80%"
              >
                <span class="absolute -top-6 left-1/2 -translate-x-1/2 bg-slate-800 text-white text-[10px] px-1.5 py-0.5 rounded">
                  4.2h
                </span>
              </div>
              <div
                class="flex-1 bg-primary/40 rounded-t-sm"
                style="height:45%"
              />
              <div
                class="flex-1 bg-primary/30 rounded-t-sm"
                style="height:20%"
              />
            </div>
            <div class="flex justify-between text-[10px] text-slate-400 mt-2 font-mono">
              <span>M</span>
              <span>T</span>
              <span>W</span>
              <span>T</span>
              <span>F</span>
              <span>S</span>
              <span>S</span>
            </div>
          </div>

          {/* Whitelist card */}
          <div class="bg-white dark:bg-[#1a1a1a] p-6 rounded-xl border border-slate-200 dark:border-white/5 shadow-sm">
            <div class="flex justify-between items-center mb-4">
              <h3 class="text-sm font-medium text-slate-500 dark:text-slate-400">
                {t("home.whitelist")}
              </h3>
              <span class="text-primary rounded p-1">
                <span class="material-symbols-outlined text-sm">add</span>
              </span>
            </div>
            <div class="space-y-3">
              <WhitelistItem
                color="blue"
                initial="V"
                name="Veritasium"
                tag="Science & Eng"
              />
              <WhitelistItem
                color="purple"
                initial="K"
                name="Kurzgesagt"
                tag="Education"
              />
              <WhitelistItem
                color="slate"
                initial="H"
                name="Huberman Lab"
                faded
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

const avatarColors: Record<string, string> = {
  blue: "bg-blue-500/20 text-blue-500",
  purple: "bg-purple-500/20 text-purple-500",
  slate: "bg-slate-500/20 text-slate-500",
};

function WhitelistItem({
  color,
  initial,
  name,
  tag,
  faded,
}: {
  color: string;
  initial: string;
  name: string;
  tag?: string;
  faded?: boolean;
}) {
  return (
    <div
      class={`flex items-center gap-3 p-2 rounded-lg bg-slate-50 dark:bg-white/5 border border-slate-100 dark:border-white/5 ${faded ? "opacity-60" : ""}`}
    >
      <div
        class={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold shrink-0 ${avatarColors[color] ?? ""}`}
      >
        {initial}
      </div>
      <div class="flex-1 min-w-0">
        <div class="text-sm font-bold text-slate-700 dark:text-slate-200 truncate">
          {name}
        </div>
        {tag && <div class="text-[10px] text-slate-400">{tag}</div>}
      </div>
      <span class="material-symbols-outlined text-slate-300 text-sm">
        check_circle
      </span>
    </div>
  );
}

export default function Home() {
  const { t } = useTranslation();

  return (
    <div class="flex flex-row h-screen w-full overflow-hidden font-display antialiased">
      {/* Left panel */}
      <div class="w-full lg:w-[45%] h-full flex flex-col justify-center relative z-20 px-8 md:px-12 lg:px-16 xl:px-20 bg-[var(--color-bg)] border-r border-slate-200 dark:border-white/5">
        {/* Logo */}
        <a
          href="/"
          class="absolute top-8 left-8 md:left-12 lg:left-16 flex items-center gap-2 no-underline text-inherit"
        >
          <span class="material-symbols-outlined text-3xl text-primary">
            timelapse
          </span>
          <h2 class="text-xl font-bold tracking-tight m-0">anti-yt</h2>
        </a>

        {/* Hero */}
        <div class="max-w-lg">
          <h1 class="text-4xl sm:text-5xl lg:text-6xl font-extrabold tracking-tight leading-[1.1] my-4">
            {t("home.heroTitle1")}
            <br />
            <span class="text-primary">{t("home.heroTitle2")}</span>.
          </h1>
          <p class="text-lg text-slate-600 dark:text-slate-400 leading-relaxed m-0">
            {t("home.heroDescription")}
          </p>

          <div class="flex flex-col gap-4 pt-8">
            <button class="flex w-full sm:w-auto items-center justify-center gap-3 rounded-xl bg-white dark:bg-[#242424] px-8 py-4 text-base font-bold text-slate-700 dark:text-white border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-[#2a2a2a] hover:border-primary/50 dark:hover:border-primary/50 transition-all shadow-lg hover:shadow-xl focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 dark:focus:ring-offset-[var(--color-bg)] cursor-pointer">
              <GoogleIcon />
              <span>{t("home.signInWithGoogle")}</span>
            </button>
          </div>
        </div>

        {/* Footer */}
        <div class="absolute bottom-8 left-8 md:left-12 lg:left-16 flex gap-6 text-xs text-slate-400">
          <a
            class="hover:text-primary transition-colors no-underline text-inherit"
            href="/privacy"
          >
            {t("home.privacy")}
          </a>
          <a
            class="hover:text-primary transition-colors no-underline text-inherit"
            href="/terms"
          >
            {t("home.terms")}
          </a>
          <span>&copy; {new Date().getFullYear()} anti-yt</span>
        </div>
      </div>

      {/* Right panel */}
      <div class="hidden lg:flex lg:w-[55%] h-full relative items-center justify-center overflow-hidden bg-slate-100 dark:bg-[#0c0c0c]">
        <div class="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_rgba(208,187,149,0.2),transparent_70%)] dark:bg-[radial-gradient(ellipse_at_top_right,_rgba(208,187,149,0.1),transparent_70%)]" />
        <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-primary/5 rounded-full blur-[100px]" />
        <DashboardPreview t={t} />
      </div>
    </div>
  );
}
