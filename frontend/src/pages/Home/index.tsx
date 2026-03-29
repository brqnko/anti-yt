import { useState, useRef, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useCanonical } from "../../hooks/useCanonical";
import { useColorMode } from "../../hooks/useColorMode";
import { useAuth } from "../../contexts/AuthContext";
import { modeIcons, modeOrder, languages } from "../../constants";
import { GoogleIcon } from "./GoogleIcon";
import { GithubIcon } from "./GithubIcon";
import { Icon } from "../../components/Icon";

function DashboardPreview({ t }: { t: (key: string) => string }) {
  return (
    <div class="animate-fade-in-up relative w-[90%] max-w-4xl bg-white dark:bg-[#151515] rounded-2xl border border-slate-200 dark:border-white/10 overflow-hidden">
      <div class="p-8 bg-slate-50/50 dark:bg-[#0f0f0f]">
        <div class="grid grid-cols-2 gap-6">
          {/* Allowance card */}
          <div class="col-span-2 bg-white dark:bg-[#1a1a1a] p-6 rounded-xl border border-slate-200 dark:border-white/5 flex items-center justify-between">
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
              <Icon name="timer" class="absolute text-primary text-xl" />
            </div>
          </div>

          {/* Chart card */}
          <div class="bg-white dark:bg-[#1a1a1a] p-6 rounded-xl border border-slate-200 dark:border-white/5">
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
          <div class="bg-white dark:bg-[#1a1a1a] p-6 rounded-xl border border-slate-200 dark:border-white/5">
            <h3 class="text-xs font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 px-3 mb-4">
              {t("home.whitelist")}
            </h3>
            <div class="flex flex-col gap-2">
              <WhitelistItem name="Veritasium" />
              <WhitelistItem name="Kurzgesagt" />
              <WhitelistItem name="Huberman Lab" />
              <div class="flex items-center gap-2 px-3 py-2 text-sm text-primary font-medium">
                <Icon name="add" class="text-[18px]" />
                <span>{t("dashboard.requestChannel")}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function WhitelistItem({ name }: { name: string }) {
  return (
    <div class="flex items-center gap-3 px-3 py-2 rounded-lg">
      <div class="size-8 rounded-full bg-gray-200 dark:bg-gray-700 shrink-0" />
      <span class="text-sm font-medium text-slate-700 dark:text-white truncate">
        {name}
      </span>
    </div>
  );
}

export default function Home() {
  const { t, i18n } = useTranslation();
  const { mode, setMode } = useColorMode();
  const { isAuthenticated, isLoading } = useAuth();
  useTitle("");
  useCanonical("/");

  const nextMode = modeOrder[(modeOrder.indexOf(mode) + 1) % modeOrder.length];
  const cycleMode = () => setMode(nextMode);

  const currentLang = i18n.language.startsWith("ja") ? "ja" : "en";
  const [langOpen, setLangOpen] = useState(false);
  const langRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (langRef.current && !langRef.current.contains(e.target as Node)) {
        setLangOpen(false);
      }
    };
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setLangOpen(false);
    };
    document.addEventListener("click", handleClick);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("click", handleClick);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  return (
    <div class="flex flex-col h-dvh w-full overflow-hidden font-display antialiased">
      {/* Header */}
      <header class="flex items-center justify-between py-4 px-4 sm:px-8 md:px-12 lg:px-16 xl:px-20 z-30 bg-slate-100 dark:bg-[#0c0c0c] border-b border-slate-200 dark:border-white/5">
        <a href="/" class="no-underline text-charcoal dark:text-white">
          <span class="text-xl font-bold tracking-tight">anti-yt</span>
        </a>
        <div class="flex items-center gap-2">
          <button
            class="w-10 h-10 flex items-center justify-center rounded-full bg-slate-200/60 dark:bg-white/10 hover:bg-slate-300/60 dark:hover:bg-white/20 transition-colors cursor-pointer text-slate-600 dark:text-slate-300"
            onClick={cycleMode}
            title={t(`common.colorMode.${nextMode}`)}
            aria-label={`${t(`common.colorMode.${mode}`)} → ${t(`common.colorMode.${nextMode}`)}`}
          >
            <Icon name={modeIcons[mode]} class="text-xl" />
          </button>
          <div class="relative" ref={langRef}>
            <button
              class="w-10 h-10 flex items-center justify-center rounded-full bg-slate-200/60 dark:bg-white/10 hover:bg-slate-300/60 dark:hover:bg-white/20 transition-colors cursor-pointer text-slate-600 dark:text-slate-300"
              onClick={() => setLangOpen(!langOpen)}
              aria-label={t("common.language")}
              aria-expanded={langOpen}
              aria-haspopup="true"
            >
              <Icon name="translate" class="text-xl" />
            </button>
            {langOpen && (
              <div role="menu" class="absolute right-0 top-full mt-2 py-2 bg-white dark:bg-[#1a1a1a] rounded-xl ring-1 ring-black/5 dark:ring-white/5 border border-slate-200 dark:border-white/10 min-w-[180px] z-50">
                {languages.map((lang) => (
                  <button
                    role="menuitem"
                    key={lang.code}
                    class={`w-full flex items-center gap-2 px-4 py-2 text-left text-sm font-medium cursor-pointer transition-colors ${
                      currentLang === lang.code
                        ? "text-primary"
                        : "text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-white/5"
                    }`}
                    onClick={() => {
                      i18n.changeLanguage(lang.code);
                      localStorage.setItem("lang", lang.code);
                      setLangOpen(false);
                    }}
                  >
                    <Icon name="check" class={`text-base ${currentLang === lang.code ? "opacity-100" : "opacity-0"}`} />
                    {lang.label}
                  </button>
                ))}
              </div>
            )}
          </div>
          <a
            class="w-10 h-10 flex items-center justify-center rounded-full bg-slate-200/60 dark:bg-white/10 hover:bg-slate-300/60 dark:hover:bg-white/20 transition-colors text-slate-600 dark:text-slate-300"
            href="https://github.com/brqnko/anti-yt"
            target="_blank"
            rel="noopener noreferrer"
            aria-label="GitHub"
          >
            <GithubIcon />
          </a>
        </div>
      </header>

      <div class="flex flex-row flex-1 w-full overflow-hidden">
        {/* Left panel */}
        <div class="w-full lg:w-[45%] h-full flex flex-col z-20 px-8 md:px-12 lg:px-16 xl:px-20 bg-[var(--color-bg)] border-r border-slate-200 dark:border-white/5">
          {/* Hero */}
          <div class="flex-1 flex flex-col justify-center max-w-lg">
            <h1 class="text-4xl sm:text-5xl lg:text-6xl font-extrabold tracking-tight leading-[1.1] my-4">
              {t("home.heroTitle1")}
              <br />
              <span class="text-primary">{t("home.heroTitle2")}</span>.
            </h1>
            <p class="text-lg text-slate-600 dark:text-slate-400 leading-relaxed m-0">
              {t("home.heroDescription")}
            </p>

            <div class="flex flex-col gap-4 pt-8">
              {isAuthenticated && !isLoading ? (
                <a
                  href="/dashboard"
                  class="flex w-full sm:w-auto items-center justify-center rounded-xl bg-primary px-8 py-4 text-base font-bold text-white hover:bg-primary/90 transition-all hover:-translate-y-0.5 no-underline cursor-pointer"
                >
                  {t("home.dashboard")}
                </a>
              ) : (
                <button
                  disabled={isLoading}
                  onClick={() => {
                    window.location.href = "/api/v1/auth/google";
                  }}
                  class="flex w-full sm:w-auto items-center justify-center gap-3 rounded-xl bg-white dark:bg-[#242424] px-8 py-4 text-base font-bold text-slate-700 dark:text-white border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-[#2a2a2a] hover:border-primary/50 dark:hover:border-primary/50 transition-all hover:-translate-y-0.5 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 dark:focus:ring-offset-[var(--color-bg)] cursor-pointer disabled:opacity-60 disabled:pointer-events-none"
                >
                  <GoogleIcon />
                  <span>{t("home.signInWithGoogle")}</span>
                </button>
              )}
            </div>
          </div>

          {/* Footer */}
          <footer class="flex items-center gap-6 text-xs text-slate-400 py-6">
            <span>&copy; {new Date().getFullYear()} anti-yt</span>
            <nav class="flex gap-4">
              <a
                class="hover:text-primary transition-colors no-underline text-inherit cursor-pointer"
                href="/terms"
              >
                {t("home.terms")}
              </a>
              <a
                class="hover:text-primary transition-colors no-underline text-inherit cursor-pointer"
                href="/privacy"
              >
                {t("home.privacy")}
              </a>
            </nav>
          </footer>
        </div>

        {/* Right panel */}
        <div class="hidden lg:flex lg:w-[55%] h-full relative items-center justify-center overflow-hidden bg-slate-100 dark:bg-[#0c0c0c]">
          <div class="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_rgba(208,187,149,0.2),transparent_70%)] dark:bg-[radial-gradient(ellipse_at_top_right,_rgba(208,187,149,0.1),transparent_70%)]" />
          <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] rounded-full bg-[radial-gradient(circle,_var(--color-primary)_0%,_transparent_70%)] opacity-5" />
          <DashboardPreview t={t} />
        </div>
      </div>
    </div>
  );
}
