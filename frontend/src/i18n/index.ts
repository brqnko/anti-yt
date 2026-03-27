import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import ja from "./locales/ja.json";
import en from "./locales/en.json";

i18n.use(initReactI18next).init({
  resources: {
    ja: { translation: ja },
    en: { translation: en },
  },
  lng:
    typeof window !== "undefined"
      ? localStorage.getItem("lang") ||
        (navigator.language.startsWith("ja") ? "ja" : "en")
      : "en",
  fallbackLng: "en",
  interpolation: {
    escapeValue: false,
  },
});

if (typeof document !== "undefined") {
  const setLang = (lng: string) => {
    document.documentElement.lang = lng.startsWith("ja") ? "ja" : "en";
  };
  setLang(i18n.language);
  i18n.on("languageChanged", setLang);
}

export default i18n;
