import { useState, useEffect, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useColorMode } from "../hooks/useColorMode";
import { useAuth } from "../contexts/AuthContext";
import { modeIcons, modeOrder, languages } from "../constants";
import { Logo } from "./Logo";

export function DashboardHeader() {
  const { t, i18n } = useTranslation();
  const { mode, setMode } = useColorMode();
  const { user, logout } = useAuth();

  const currentLang = i18n.language.startsWith("ja") ? "ja" : "en";
  const [langOpen, setLangOpen] = useState(false);
  const langRef = useRef<HTMLDivElement>(null);

  const nextMode = modeOrder[(modeOrder.indexOf(mode) + 1) % modeOrder.length];
  const cycleMode = () => setMode(nextMode);

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
    <header class="sticky top-0 z-50 flex items-center justify-between whitespace-nowrap border-b border-solid border-border-light dark:border-border-dark bg-background-light/95 dark:bg-background-dark/95 backdrop-blur-md px-6 py-3">
      <Logo />

      <div class="flex items-center gap-4">
        <button
          class="size-9 flex items-center justify-center rounded-full hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer text-text-muted-light dark:text-text-muted-dark transition-colors"
          onClick={cycleMode}
          title={t(`common.colorMode.${nextMode}`)}
        >
          <span class="material-symbols-outlined">
            {modeIcons[mode]}
          </span>
        </button>
        <div class="relative" ref={langRef}>
          <button
            class="size-9 flex items-center justify-center rounded-full hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer text-text-muted-light dark:text-text-muted-dark transition-colors"
            onClick={() => setLangOpen(!langOpen)}
            aria-label={t("legal.languageSelect")}
            aria-expanded={langOpen}
            aria-haspopup="true"
          >
            <span class="material-symbols-outlined">translate</span>
          </button>
          {langOpen && (
            <div role="menu" class="absolute right-0 top-full mt-2 py-2 bg-card-light dark:bg-card-dark rounded-xl shadow-xl border border-border-light dark:border-border-dark min-w-[180px] z-50">
              {languages.map((lang) => (
                <button
                  role="menuitem"
                  key={lang.code}
                  class={`w-full flex items-center gap-2 px-4 py-2 text-left text-sm font-medium cursor-pointer transition-colors bg-transparent border-none ${
                    currentLang === lang.code
                      ? "text-primary"
                      : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5"
                  }`}
                  onClick={() => {
                    i18n.changeLanguage(lang.code);
                    localStorage.setItem("lang", lang.code);
                    setLangOpen(false);
                  }}
                >
                  <span class={`material-symbols-outlined text-base ${currentLang === lang.code ? "opacity-100" : "opacity-0"}`}>
                    check
                  </span>
                  {lang.label}
                </button>
              ))}
            </div>
          )}
        </div>
        <a
          href="/profile"
          class="size-9 flex items-center justify-center rounded-full bg-primary/10 ring-2 ring-primary/20 cursor-pointer text-primary font-bold text-sm no-underline"
          title={t("profile.pageTitle")}
        >
          {user?.display_name?.charAt(0)?.toUpperCase() ?? "?"}
        </a>
      </div>
    </header>
  );
}
