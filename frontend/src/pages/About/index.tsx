import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useCanonical } from "../../hooks/useCanonical";
import { HeaderControls } from "../../components/HeaderControls";

const principles = ["feed", "time", "analytics"] as const;
const featureKeys = ["feed", "limit", "analytics"] as const;

function GoogleIcon() {
  return (
    <svg
      class="w-5 h-5 shrink-0"
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
        fill="#4285F4"
      />
      <path
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        fill="#34A853"
      />
      <path
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
        fill="#FBBC05"
      />
      <path
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        fill="#EA4335"
      />
    </svg>
  );
}

function GoogleSignInButton({ label }: { label: string }) {
  return (
    <button
      type="button"
      onClick={() => {
        window.location.href = "/api/v1/auth/google";
      }}
      class="flex items-center justify-center gap-3 rounded-xl bg-white dark:bg-[#242424] px-8 py-3 text-sm font-bold text-slate-700 dark:text-white border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-[#2a2a2a] hover:border-primary/50 dark:hover:border-primary/50 transition-colors cursor-pointer"
    >
      <GoogleIcon />
      {label}
    </button>
  );
}

export default function About() {
  const { t } = useTranslation();
  useTitle(t("about.pageTitle"));
  useCanonical("/about");

  return (
    <div class="min-h-dvh bg-background-light dark:bg-background-dark text-charcoal dark:text-white font-display">
      <header class="border-b border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark sticky top-0 z-50">
        <div class="max-w-7xl mx-auto px-6 md:px-8 py-4 flex items-center justify-between">
          <a href="/" class="no-underline text-inherit">
            <span class="text-xl font-bold tracking-tight">anti-yt</span>
          </a>
          <HeaderControls />
        </div>
      </header>

      <main class="pt-8 md:pt-10">
        {/* Hero */}
        <section class="max-w-7xl mx-auto px-6 md:px-8 mb-14 md:mb-20">
          <div class="grid lg:grid-cols-12 gap-8 lg:gap-16 items-center">
            <div class="lg:col-span-7 space-y-6 md:space-y-8">
              <h1 class="text-4xl md:text-5xl xl:text-6xl font-extrabold text-charcoal dark:text-white leading-[1.1] tracking-tighter">
                {t("about.hero.title1")}
                <br />
                {t("about.hero.title2")}
              </h1>
              <p class="text-lg md:text-xl text-taupe dark:text-text-muted-dark leading-relaxed max-w-xl font-medium">
                {t("about.hero.description")}
              </p>
            </div>
            <div class="lg:col-span-5 relative mt-6 lg:mt-0">
              <div class="aspect-square rounded-2xl overflow-hidden bg-background-light border border-border-light shadow-sm relative flex items-center justify-center">
                <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
                  <div class="w-72 h-72 bg-green-400/35 rounded-full blur-3xl translate-x-16" />
                </div>
                <img
                  src="/about-preview.png"
                  alt=""
                  class="w-full h-full object-contain p-8 relative"
                />
              </div>
            </div>
          </div>
        </section>

        {/* Editorial statement */}
        <section class="max-w-7xl mx-auto px-6 md:px-8 mb-14 md:mb-20">
          <div class="max-w-4xl border-l-4 border-primary pl-6 md:pl-10 py-4 md:py-6">
            <h2 class="text-2xl md:text-4xl font-bold text-charcoal dark:text-white mb-6 md:mb-8 leading-tight">
              {t("about.statement.heading")}
            </h2>
            <div class="space-y-4 md:space-y-6 text-base md:text-xl text-taupe dark:text-text-muted-dark leading-relaxed">
              <p>{t("about.statement.body1")}</p>
              <p class="font-bold text-charcoal dark:text-white">
                {t("about.statement.body2")}
              </p>
            </div>
          </div>
        </section>

        {/* Principles */}
        <section class="max-w-7xl mx-auto px-6 md:px-8 overflow-hidden mb-14 md:mb-20">
          <div class="flex flex-col gap-12 md:gap-20">
            {principles.map((key, idx) => {
              const num = String(idx + 1).padStart(2, "0");
              const isOdd = idx % 2 === 1;
              return (
                <div key={key} class="grid md:grid-cols-12 items-end">
                  {!isOdd ? (
                    <>
                      <div class="md:col-span-2 hidden md:block">
                        <span class="text-9xl font-black text-primary/40 dark:text-primary/30 tracking-tighter leading-none select-none">
                          {num}
                        </span>
                      </div>
                      <div
                        class={`md:col-span-6 space-y-4 md:space-y-6 ${
                          idx === 0 ? "md:col-start-4" : "md:col-start-5"
                        }`}
                      >
                        <h3 class="text-2xl md:text-4xl font-bold text-charcoal dark:text-white leading-tight">
                          {t(`about.principles.${key}.title`)}
                        </h3>
                        <p class="text-base md:text-lg text-taupe dark:text-text-muted-dark leading-relaxed">
                          {t(`about.principles.${key}.body`)}
                        </p>
                      </div>
                    </>
                  ) : (
                    <>
                      <div class="md:col-span-6 md:col-start-3 space-y-4 md:space-y-6 md:text-right order-2 md:order-1">
                        <h3 class="text-2xl md:text-4xl font-bold text-charcoal dark:text-white leading-tight">
                          {t(`about.principles.${key}.title`)}
                        </h3>
                        <p class="text-base md:text-lg text-taupe dark:text-text-muted-dark leading-relaxed md:ml-auto md:max-w-md">
                          {t(`about.principles.${key}.body`)}
                        </p>
                      </div>
                      <div class="md:col-span-2 md:col-start-10 hidden md:block order-1 md:order-2">
                        <span class="text-9xl font-black text-primary/40 dark:text-primary/30 tracking-tighter leading-none select-none">
                          {num}
                        </span>
                      </div>
                    </>
                  )}
                </div>
              );
            })}
          </div>
        </section>

        {/* Feature spec */}
        <section class="bg-card-light dark:bg-card-dark px-6 md:px-8 py-12 md:py-16">
          <div class="max-w-7xl mx-auto">
            <div class="mb-8 md:mb-10">
              <h3 class="text-2xl md:text-4xl font-bold text-charcoal dark:text-white max-w-2xl leading-tight">
                {t("about.features.heading")}
              </h3>
            </div>
            <div class="grid md:grid-cols-2 gap-x-10 md:gap-x-16 gap-y-8 md:gap-y-10">
              {featureKeys.map((key, idx) => (
                <div
                  key={key}
                  class={`space-y-3 md:space-y-4 ${idx === 1 ? "md:pt-8" : ""}`}
                >
                  <h4 class="text-xl md:text-2xl font-bold text-charcoal dark:text-white">
                    {t(`about.features.${key}.title`)}
                  </h4>
                  <p class="text-taupe dark:text-text-muted-dark text-base md:text-lg leading-relaxed">
                    {t(`about.features.${key}.body`)}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Final CTA */}
        <section class="px-6 md:px-8 text-center max-w-5xl mx-auto py-14 md:py-20">
          <div class="space-y-8 md:space-y-10">
            <p class="text-xl md:text-3xl font-bold text-charcoal dark:text-white max-w-2xl mx-auto leading-snug">
              {t("about.finalCta.description")}
            </p>
            <div class="flex flex-col items-center gap-8">
              <GoogleSignInButton label={t("common.signInWithGoogle")} />
              <p class="text-xs text-text-muted-light dark:text-text-muted-dark leading-relaxed text-center max-w-md">
                {t("common.oidcNoticeNoPii")}{" "}
                {t("common.privacyPolicyConsentBefore")}
                <a
                  href="/terms"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="underline hover:text-charcoal dark:hover:text-white transition-colors"
                >
                  {t("common.termsLink")}
                </a>
                {t("common.privacyPolicyConsentMiddle")}
                <a
                  href="/privacy"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="underline hover:text-charcoal dark:hover:text-white transition-colors"
                >
                  {t("common.privacyPolicyLink")}
                </a>
                {t("common.privacyPolicyConsentAfter")}
              </p>
            </div>
          </div>
        </section>

        {/* Footer */}
        <footer class="border-t border-border-light dark:border-border-dark">
          <div class="max-w-7xl mx-auto px-6 md:px-8 py-6 flex flex-col md:flex-row justify-between items-center gap-4 text-sm text-taupe dark:text-text-muted-dark">
            <div class="flex items-center gap-6">
              <a
                href="/terms"
                class="font-medium no-underline text-inherit hover:text-primary transition-colors"
              >
                {t("legal.navTerms")}
              </a>
              <a
                href="/privacy"
                class="font-medium no-underline text-inherit hover:text-primary transition-colors"
              >
                {t("legal.navPrivacy")}
              </a>
            </div>
            <span>&copy; {new Date().getFullYear()} anti-yt</span>
          </div>
        </footer>
      </main>
    </div>
  );
}
