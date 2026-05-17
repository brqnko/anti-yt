import { useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { SITE_URL } from "../constants";
import { useTitle } from "./useTitle";
import { useCanonical } from "./useCanonical";

interface UseMetaParams {
  title?: string;
  description?: string;
  canonicalPath?: string;
  ogImage?: string;
  ogType?: "website" | "article" | "video.other";
}

const OG_LOCALE_MAP: Record<string, string> = {
  en: "en_US",
  ja: "ja_JP",
  zh: "zh_CN",
};

function setMeta(selector: string, attr: "name" | "property", key: string, value: string) {
  let el = document.head.querySelector<HTMLMetaElement>(selector);
  if (!el) {
    el = document.createElement("meta");
    el.setAttribute(attr, key);
    document.head.appendChild(el);
  }
  el.content = value;
}

export function useMeta(params: UseMetaParams) {
  const { title, description, canonicalPath, ogImage, ogType = "website" } = params;
  const { i18n } = useTranslation();
  const lang = i18n.resolvedLanguage || i18n.language || "en";
  const ogLocale = OG_LOCALE_MAP[lang.slice(0, 2)] ?? "en_US";

  useTitle(title ?? "");
  useCanonical(canonicalPath ?? (typeof window === "undefined" ? "/" : window.location.pathname));

  useEffect(() => {
    if (typeof document === "undefined") return;

    const fullTitle = title ? `${title} | anti-yt` : "anti-yt";
    const finalImage = ogImage ?? `${SITE_URL}/og-image.png`;
    const finalUrl = `${SITE_URL}${canonicalPath ?? (typeof window === "undefined" ? "/" : window.location.pathname)}`;

    if (description) {
      setMeta('meta[name="description"]', "name", "description", description);
      setMeta('meta[property="og:description"]', "property", "og:description", description);
      setMeta('meta[name="twitter:description"]', "name", "twitter:description", description);
    }

    setMeta('meta[property="og:title"]', "property", "og:title", fullTitle);
    setMeta('meta[name="twitter:title"]', "name", "twitter:title", fullTitle);

    setMeta('meta[property="og:url"]', "property", "og:url", finalUrl);
    setMeta('meta[property="og:image"]', "property", "og:image", finalImage);
    setMeta('meta[name="twitter:image"]', "name", "twitter:image", finalImage);

    setMeta('meta[property="og:type"]', "property", "og:type", ogType);
    setMeta('meta[property="og:locale"]', "property", "og:locale", ogLocale);
  }, [title, description, canonicalPath, ogImage, ogType, ogLocale]);
}
