import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import type { ComponentChildren } from "preact";
import { HeaderControls } from "./HeaderControls";

const navLinks = [
  { href: "/terms", labelKey: "legal.navTerms" },
  { href: "/privacy", labelKey: "legal.navPrivacy" },
];

export function LegalLayout({ children }: { children: ComponentChildren }) {
  const { t } = useTranslation();
  const { url } = useLocation();

  return (
    <div class="min-h-dvh bg-[#fbfaf9] dark:bg-[#1d1a15] text-[#171511] dark:text-[#fbfaf9] font-display">
      {/* Header */}
      <header class="border-b border-[#f0eeea] dark:border-[#3d372e] bg-[#fbfaf9] dark:bg-[#1d1a15] sticky top-0 z-50">
        <div class="max-w-[1000px] mx-auto px-6 py-4 flex items-center justify-between">
          <a href="/" class="no-underline text-inherit">
            <span class="text-xl font-bold tracking-tight">anti-yt</span>
          </a>
          <div class="flex items-center gap-6">
            <nav class="hidden sm:flex items-center gap-6">
              {navLinks.map((link) => (
                <a
                  key={link.href}
                  href={link.href}
                  class={`text-sm font-bold no-underline ${
                    url === link.href
                      ? "text-primary"
                      : "text-[#847862] dark:text-[#a89d89] hover:text-primary"
                  }`}
                >
                  {t(link.labelKey)}
                </a>
              ))}
            </nav>
            <HeaderControls />
          </div>
        </div>
      </header>

      <main class="max-w-[1000px] mx-auto px-6 py-16 md:py-24 text-left">
        {children}

        {/* Footer */}
        <footer class="max-w-4xl mx-auto py-16 border-t border-[#e1ddd6] dark:border-[#3d372e] flex flex-col md:flex-row justify-between items-center gap-6 text-sm text-[#847862]">
          <nav class="flex items-center gap-6">
            {navLinks.map((link) => (
              <a
                key={link.href}
                href={link.href}
                class={`font-medium no-underline ${
                  url === link.href
                    ? "text-primary"
                    : "text-inherit hover:text-primary"
                }`}
              >
                {t(link.labelKey)}
              </a>
            ))}
          </nav>
          <span>&copy; {new Date().getFullYear()} anti-yt</span>
        </footer>
      </main>
    </div>
  );
}
