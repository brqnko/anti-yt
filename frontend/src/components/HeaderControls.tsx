import { useState, useRef, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { Icon } from "./Icon";
import { useColorMode } from "../hooks/useColorMode";
import { modeIcons, modeOrder, languages } from "../constants";

export function HeaderControls() {
  const { t, i18n } = useTranslation();
  const { mode, setMode } = useColorMode();

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
    <div class="flex items-center gap-2">
      <button
        type="button"
        class="w-10 h-10 flex items-center justify-center rounded-full bg-slate-200/60 dark:bg-white/10 hover:bg-slate-300/60 dark:hover:bg-white/20 transition-colors cursor-pointer text-slate-600 dark:text-slate-300"
        onClick={cycleMode}
        title={t(`common.colorMode.${nextMode}`)}
        aria-label={`${t(`common.colorMode.${mode}`)} → ${t(`common.colorMode.${nextMode}`)}`}
      >
        <Icon name={modeIcons[mode]} class="text-xl" />
      </button>
      <div class="relative" ref={langRef}>
        <button
          type="button"
          class="w-10 h-10 flex items-center justify-center rounded-full bg-slate-200/60 dark:bg-white/10 hover:bg-slate-300/60 dark:hover:bg-white/20 transition-colors cursor-pointer text-slate-600 dark:text-slate-300"
          onClick={() => setLangOpen(!langOpen)}
          aria-label={t("legal.languageSelect")}
          aria-expanded={langOpen}
          aria-haspopup="true"
        >
          <Icon name="translate" class="text-xl" />
        </button>
        {langOpen && (
          <div
            role="menu"
            class="absolute right-0 top-full mt-2 py-2 bg-white dark:bg-[#1a1a1a] rounded-xl ring-1 ring-black/5 dark:ring-white/5 border border-slate-200 dark:border-white/10 min-w-[180px] z-50"
          >
            {languages.map((lang) => (
              <button
                role="menuitem"
                type="button"
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
                <Icon
                  name="check"
                  class={`text-base ${currentLang === lang.code ? "opacity-100" : "opacity-0"}`}
                />
                {lang.label}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
