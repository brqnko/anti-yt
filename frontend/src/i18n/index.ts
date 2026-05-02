import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import ja from "./locales/ja.json";
import en from "./locales/en.json";
import zh from "./locales/zh.json";

function detectLanguage(): string {
  if (typeof window === "undefined") return "en";
  const saved = localStorage.getItem("lang");
  if (saved) return saved;
  const nav = navigator.language;
  if (nav.startsWith("ja")) return "ja";
  if (nav.startsWith("zh")) return "zh";
  return "en";
}

i18n.use(initReactI18next).init({
  resources: {
    ja: { translation: ja },
    en: { translation: en },
    zh: { translation: zh },
  },
  lng: detectLanguage(),
  fallbackLng: "en",
  interpolation: {
    escapeValue: false,
  },
});

if (typeof document !== "undefined") {
  const setLang = (lng: string) => {
    document.documentElement.lang = lng;
  };
  setLang(i18n.language);
  i18n.on("languageChanged", setLang);
}

export default i18n;
