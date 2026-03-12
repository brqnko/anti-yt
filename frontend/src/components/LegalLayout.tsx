import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import type { ComponentChildren } from "preact";

const languages = [
  { code: "ja", label: "日本語" },
  { code: "en", label: "English" },
];

const navLinks = [
  { href: "/terms", labelKey: "legal.navTerms" },
  { href: "/privacy", labelKey: "legal.navPrivacy" },
];

export function LegalLayout({ children }: { children: ComponentChildren }) {
  const { t, i18n } = useTranslation();
  const { url } = useLocation();

  return (
    <div class="min-h-screen bg-[#fbfaf9] dark:bg-[#1d1a15] text-[#171511] dark:text-[#fbfaf9] font-display">
      {/* Header */}
      <header class="border-b border-[#f0eeea] dark:border-[#3d372e] bg-[#fbfaf9]/80 dark:bg-[#1d1a15]/80 backdrop-blur-md sticky top-0 z-50">
        <div class="max-w-[1000px] mx-auto px-6 py-4 flex items-center justify-between">
          <a href="/" class="flex items-center gap-4 no-underline text-inherit">
            <span class="material-symbols-outlined text-3xl text-primary">
              timelapse
            </span>
            <h2 class="text-xl font-bold leading-tight tracking-tight m-0">
              anti-yt
            </h2>
          </a>
          <nav class="flex items-center gap-6">
            {navLinks.map((link) => (
              <a
                key={link.href}
                href={link.href}
                class={`text-sm font-bold no-underline transition-colors ${
                  url === link.href
                    ? "text-primary"
                    : "text-[#847862] dark:text-[#a89d89] hover:text-primary"
                }`}
              >
                {t(link.labelKey)}
              </a>
            ))}
          </nav>
        </div>
      </header>

      <main class="max-w-[1000px] mx-auto px-6 py-16 md:py-24 text-left">
        {children}

        {/* Footer */}
        <footer class="max-w-4xl mx-auto py-16 border-t border-[#e1ddd6] dark:border-[#3d372e] flex flex-col md:flex-row justify-between items-center gap-6 text-sm text-[#847862]">
          <div class="flex items-center gap-6">
            {navLinks.map((link) => (
              <a
                key={link.href}
                href={link.href}
                class={`font-medium no-underline transition-colors ${
                  url === link.href
                    ? "text-primary"
                    : "text-inherit hover:text-primary"
                }`}
              >
                {t(link.labelKey)}
              </a>
            ))}
          </div>
          <div class="flex items-center gap-4">
            <select
              class="text-sm font-semibold text-[#847862] dark:text-[#a89d89] bg-[#f0eeea] dark:bg-[#3d372e] border-none rounded-full px-4 py-1.5 cursor-pointer outline-none"
              aria-label={t("legal.languageSelect")}
              value={i18n.language}
              onChange={(e) => {
                const lang = (e.target as HTMLSelectElement).value;
                i18n.changeLanguage(lang);
                localStorage.setItem("lang", lang);
              }}
            >
              {languages.map((lang) => (
                <option key={lang.code} value={lang.code}>
                  {lang.label}
                </option>
              ))}
            </select>
            <span>&copy; {new Date().getFullYear()} anti-yt</span>
          </div>
        </footer>
      </main>
    </div>
  );
}
