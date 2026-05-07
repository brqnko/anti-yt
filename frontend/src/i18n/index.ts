import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import en from "./locales/en.json";

function detectLanguage(): string {
  if (typeof window === "undefined") return "en";
  const saved = localStorage.getItem("lang");
  if (saved) return saved;
  const nav = navigator.language;
  if (nav.startsWith("ja")) return "ja";
  if (nav.startsWith("zh")) return "zh";
  return "en";
}

const lng = detectLanguage();

i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
  },
  lng,
  fallbackLng: "en",
  interpolation: {
    escapeValue: false,
  },
});

async function loadLocale(target: string): Promise<void> {
  if (i18n.hasResourceBundle(target, "translation")) return;
  let mod: { default: Record<string, unknown> } | null = null;
  try {
    if (target === "ja") {
      mod = await import("./locales/ja.json");
    } else if (target === "zh") {
      mod = await import("./locales/zh.json");
    }
  } catch {
    return;
  }
  if (mod) {
    i18n.addResourceBundle(target, "translation", mod.default, true, true);
  }
}

if (typeof window !== "undefined" && lng !== "en") {
  void loadLocale(lng).then(() => {
    void i18n.changeLanguage(lng);
  });
}

if (typeof window !== "undefined") {
  i18n.on("languageChanged", (next) => {
    if (next !== "en") void loadLocale(next);
  });
}

if (typeof document !== "undefined") {
  const setLang = (next: string) => {
    document.documentElement.lang = next;
  };
  setLang(i18n.language);
  i18n.on("languageChanged", setLang);
}

export default i18n;
