import { useTranslation, Trans } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { useCanonical } from "../../hooks/useCanonical";
import { HeaderControls } from "../../components/HeaderControls";
import { Reveal } from "../../components/Reveal";

const principles = ["feed", "time", "analytics"] as const;

function MockWindow({ children }: { children: preact.ComponentChildren }) {
  return (
    <div
      class="rounded-2xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark shadow-xl overflow-hidden select-none"
      aria-hidden="true"
    >
      <div class="flex items-center gap-3 px-4 py-3 border-b border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark">
        <div class="flex items-center gap-1.5">
          <span class="size-3 rounded-full bg-red-400" />
          <span class="size-3 rounded-full bg-yellow-400" />
          <span class="size-3 rounded-full bg-green-400" />
        </div>
      </div>
      <div class="bg-background-light dark:bg-background-dark">{children}</div>
    </div>
  );
}

const FEED_CHANNELS_BY_LANG: Record<
  string,
  Array<{ name: string; id: string; subs: string }>
> = {
  ja: [
    { name: "中田敦彦のYouTube大学", id: "@nakataatsuhiko", subs: "5.6M" },
    { name: "両学長 リベラルアーツ大学", id: "@ryogakucho", subs: "2.5M" },
    { name: "PIVOT 公式チャンネル", id: "@pivot00", subs: "1.5M" },
  ],
  en: [
    { name: "Kurzgesagt – In a Nutshell", id: "@kurzgesagt", subs: "23M" },
    { name: "Veritasium", id: "@veritasium", subs: "17M" },
    { name: "3Blue1Brown", id: "@3blue1brown", subs: "7.2M" },
  ],
  zh: [
    { name: "老高與小茉 Mr & Mrs Gao", id: "@oldgao", subs: "7.8M" },
    { name: "一席 YiXi", id: "@yixitalks", subs: "1.5M" },
    { name: "李永乐老师", id: "@liyongle", subs: "1.2M" },
  ],
};

function FeedMock({ t }: { t: (key: string) => string }) {
  const { i18n } = useTranslation();
  const lang = (i18n.resolvedLanguage || i18n.language || "en").slice(0, 2);
  const channels = FEED_CHANNELS_BY_LANG[lang] ?? FEED_CHANNELS_BY_LANG.en;
  return (
    <MockWindow>
      <div class="flex flex-col gap-6 p-6 md:p-8 bg-background-light dark:bg-background-dark">
        <h3 class="text-xl md:text-2xl font-bold text-charcoal dark:text-white tracking-tight">
          {t("channels.pageTitle")}
        </h3>
        <div class="rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark overflow-hidden">
          <div class="p-5 flex flex-col gap-3">
            {channels.map((ch) => (
              <div
                key={ch.id}
                class="flex items-center gap-4 p-3 rounded-lg bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark"
              >
                <div class="flex flex-col grow min-w-0">
                  <span class="font-bold truncate text-charcoal dark:text-white text-sm">
                    {ch.name}
                  </span>
                  <span class="text-xs text-text-muted-light dark:text-text-muted-dark truncate">
                    {ch.id} · {ch.subs} {t("channelDetail.subscribers")}
                  </span>
                </div>
                <span class="size-8 flex items-center justify-center rounded-full text-text-muted-light dark:text-text-muted-dark shrink-0 text-xl leading-none">
                  ×
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </MockWindow>
  );
}

function AnalyticsMock({ t }: { t: (key: string) => string }) {
  const stats = [
    { label: t("analytics.timeWasted"), value: "5h 42m" },
    { label: t("analytics.dailyAverage"), value: "48m" },
    { label: t("analytics.totalVideos"), value: `27${t("analytics.totalVideosUnit")}` },
  ];
  const bars = [55, 70, 40, 95, 25, 80, 60];
  const days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
  return (
    <MockWindow>
      <div class="flex flex-col gap-6 p-6 md:p-8 bg-background-light dark:bg-background-dark">
        <h3 class="text-xl md:text-2xl font-bold text-charcoal dark:text-white tracking-tight">
          {t("analytics.title")}
        </h3>

        <div class="grid grid-cols-3 gap-3">
          {stats.map((s) => (
            <div
              key={s.label}
              class="flex flex-col gap-1 rounded-xl p-3 border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark"
            >
              <p class="text-[10px] font-medium uppercase tracking-wider text-text-muted-light dark:text-text-muted-dark truncate">
                {s.label}
              </p>
              <p class="text-base font-bold text-charcoal dark:text-white">
                {s.value}
              </p>
            </div>
          ))}
        </div>

        <div class="rounded-xl border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark overflow-hidden">
          <div class="p-4 border-b border-border-light dark:border-border-dark">
            <span class="text-sm font-bold text-charcoal dark:text-white">
              {t("analytics.weeklyUsageTrends")}
            </span>
          </div>
          <div class="p-4">
            <div class="flex items-end justify-between gap-2 h-32">
              {bars.map((h, i) => (
                <div
                  key={days[i]}
                  class="flex flex-col items-center gap-2 h-full justify-end flex-1"
                >
                  <div
                    class="w-full max-w-[24px] rounded-t-md bg-primary/80"
                    style={{ height: `${h}%` }}
                  />
                  <p class="text-[10px] font-bold tracking-wider text-text-muted-light dark:text-text-muted-dark">
                    {days[i]}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </MockWindow>
  );
}

function ScreenTimeMock({ t }: { t: (key: string) => string }) {
  const ranges = [
    { start: 7, end: 9 },
    { start: 19, end: 22 },
  ];
  return (
    <div
      class="rounded-2xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark shadow-xl overflow-hidden select-none"
      aria-hidden="true"
    >
      {/* Window chrome */}
      <div class="flex items-center gap-3 px-4 py-3 border-b border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark">
        <div class="flex items-center gap-1.5">
          <span class="size-3 rounded-full bg-red-400" />
          <span class="size-3 rounded-full bg-yellow-400" />
          <span class="size-3 rounded-full bg-green-400" />
        </div>
      </div>

      <div class="flex flex-col gap-6 p-6 md:p-8 bg-background-light dark:bg-background-dark">
        <h3 class="text-xl md:text-2xl font-bold text-charcoal dark:text-white tracking-tight">
          {t("restrictions.timeConstraints")}
        </h3>

      {/* Permitted Hours card */}
      <div class="rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark overflow-hidden">
        <div class="p-5 flex flex-col gap-5">
          <div class="flex justify-between items-center flex-wrap gap-2">
            <label class="text-base font-semibold text-charcoal dark:text-white">
              {t("restrictions.permittedHours")}
            </label>
            <span class="text-xs font-medium text-text-muted-light dark:text-text-muted-dark">
              {t("restrictions.permittedHoursDesc")}
            </span>
          </div>

          <div class="flex flex-col gap-4">
            {ranges.map(({ start, end }) => {
              const startPct = (start / 24) * 100;
              const endPct = (end / 24) * 100;
              const startLabel = `${String(start).padStart(2, "0")}:00`;
              const endLabel = `${String(end).padStart(2, "0")}:00`;
              return (
                <div
                  key={startLabel}
                  class="flex flex-col gap-4 p-4 rounded-xl bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark"
                >
                  <div class="flex items-center justify-between">
                    <div class="px-2.5 py-1 bg-primary/10 text-primary text-xs font-bold rounded border border-primary/20">
                      {startLabel} - {endLabel}
                    </div>
                  </div>
                  <div class="relative pt-2 px-1">
                    <div class="flex justify-between text-[10px] font-medium text-text-muted-light dark:text-text-muted-dark mb-2">
                      <span>00:00</span>
                      <span class="hidden sm:inline">06:00</span>
                      <span>12:00</span>
                      <span class="hidden sm:inline">18:00</span>
                      <span>24:00</span>
                    </div>
                    <div class="h-1.5 w-full bg-border-light dark:bg-border-dark rounded-full relative">
                      <div
                        class="absolute h-full bg-primary rounded-full"
                        style={{ left: `${startPct}%`, width: `${endPct - startPct}%` }}
                      />
                      <div
                        class="absolute top-1/2 -mt-2 -ml-2 size-4 bg-white dark:bg-card-dark border-2 border-primary rounded-full"
                        style={{ left: `${startPct}%` }}
                      />
                      <div
                        class="absolute top-1/2 -mt-2 -ml-2 size-4 bg-white dark:bg-card-dark border-2 border-primary rounded-full"
                        style={{ left: `${endPct}%` }}
                      />
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          <div class="flex items-center justify-center gap-2 w-full py-3 border border-dashed border-border-light dark:border-border-dark rounded-lg text-text-muted-light dark:text-text-muted-dark">
            <span class="text-xl leading-none">+</span>
            <span class="text-sm font-bold">
              {t("restrictions.addTimeRange")}
            </span>
          </div>
        </div>
      </div>

      {/* Daily Cap card */}
      <div class="rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark overflow-hidden">
        <div class="p-5 flex flex-col gap-4">
          <label class="text-base font-semibold text-charcoal dark:text-white">
            {t("restrictions.dailyCapLimit")}
          </label>
          <div class="flex flex-wrap gap-4 items-center">
            <span class="relative inline-flex h-7 w-12 shrink-0 items-center rounded-full bg-primary">
              <span class="inline-block size-5 rounded-full bg-white translate-x-6" />
            </span>
            <span class="text-sm font-medium text-text-muted-light dark:text-text-muted-dark">
              {t("restrictions.enableLimit")}
            </span>
            <div class="w-px h-6 bg-border-light dark:bg-border-dark" />
            <div class="flex flex-wrap gap-3 items-center">
              <div class="relative">
                <span class="block w-20 pl-4 pr-8 py-3 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg text-center font-bold text-lg text-charcoal dark:text-white">
                  2
                </span>
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-text-muted-light dark:text-text-muted-dark font-medium">
                  {t("restrictions.hr")}
                </span>
              </div>
              <span class="text-text-muted-light dark:text-text-muted-dark font-bold">:</span>
              <div class="relative">
                <span class="block w-20 pl-4 pr-8 py-3 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg text-center font-bold text-lg text-charcoal dark:text-white">
                  00
                </span>
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-text-muted-light dark:text-text-muted-dark font-medium">
                  {t("restrictions.min")}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
      </div>
    </div>
  );
}

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
      class="flex items-center justify-center gap-3 rounded-xl bg-white dark:bg-[#242424] px-8 py-3 text-sm font-bold text-slate-700 dark:text-white border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-[#2a2a2a] hover:border-primary/50 dark:hover:border-primary/50 cursor-pointer"
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
          <div class="grid lg:grid-cols-[1.05fr_1fr] items-center gap-12 lg:gap-20">
            <Reveal class="flex flex-col gap-6 md:gap-8 lg:pl-16 xl:pl-24">
              <h1 class="text-5xl md:text-6xl xl:text-7xl font-black text-charcoal dark:text-white leading-[1.05] tracking-tighter">
                {t("about.hero.title1")}
                <br />
                {t("about.hero.title2")}
              </h1>
              <p class="text-lg md:text-xl text-taupe dark:text-text-muted-dark leading-relaxed font-medium max-w-md">
                {t("about.hero.description")}
              </p>
            </Reveal>
            <Reveal
              as="ul"
              delay={150}
              class="flex flex-col gap-5 md:gap-6 list-disc pl-7 text-xl md:text-2xl font-bold text-charcoal dark:text-white marker:text-primary leading-snug"
            >
              <li>{t("about.hero.feature1")}</li>
              <li>{t("about.hero.feature2")}</li>
              <li>{t("about.hero.feature3")}</li>
            </Reveal>
          </div>
        </section>

        {/* Editorial statement */}
        <section class="max-w-7xl mx-auto px-6 md:px-8 mb-14 md:mb-20">
          <div class="lg:pl-16 xl:pl-24">
            <Reveal class="max-w-2xl border-l-4 border-primary pl-6 md:pl-10 py-4 md:py-6">
              <h2 class="text-2xl md:text-4xl font-bold text-charcoal dark:text-white mb-6 md:mb-8 leading-tight">
                {t("about.statement.heading")}
              </h2>
              <div class="space-y-4 md:space-y-8 text-base md:text-xl text-taupe dark:text-text-muted-dark leading-relaxed">
                <p>{t("about.statement.body1")}</p>
                <p class="text-xl md:text-3xl font-black text-charcoal dark:text-white leading-snug tracking-tight">
                  {t("about.statement.body2")}
                </p>
              </div>
            </Reveal>
          </div>
        </section>

        {/* Principles */}
        <section class="max-w-7xl mx-auto px-6 md:px-8 overflow-x-clip mb-14 md:mb-20 pb-16 md:pb-32">
          <div class="lg:pl-16 xl:pl-24 flex flex-col gap-12 md:gap-20">
            {principles.map((key, idx) => {
              const num = String(idx + 1).padStart(2, "0");

              const blurClass =
                key === "feed"
                  ? "bg-red-400/25"
                  : key === "time"
                    ? "bg-green-400/25"
                    : "bg-blue-400/25";
              const mockNode =
                key === "feed" ? (
                  <FeedMock t={t} />
                ) : key === "time" ? (
                  <ScreenTimeMock t={t} />
                ) : (
                  <AnalyticsMock t={t} />
                );
              const mockBlock = (
                <div class="relative">
                  <div
                    aria-hidden="true"
                    class="pointer-events-none absolute inset-0 flex items-center justify-center"
                  >
                    <div
                      class={`w-[750px] h-[450px] rounded-[100%] blur-3xl ${blurClass}`}
                    />
                  </div>
                  <div class="relative md:origin-top-left md:scale-[0.8] md:w-[125%] md:[margin-bottom:-20%]">
                    {mockNode}
                  </div>
                </div>
              );

              if (key === "time") {
                return (
                  <Reveal
                    key={key}
                    class="md:grid md:grid-cols-12 md:items-start md:gap-y-8 flex flex-col gap-8 md:mt-32 md:mb-16"
                  >
                    <div class="md:col-start-2 md:col-span-6 md:row-start-1 space-y-4 md:space-y-6 md:text-right order-2 md:order-1">
                      <h3 class="text-2xl md:text-4xl font-bold text-charcoal dark:text-white leading-tight">
                        {t(`about.principles.${key}.title`)}
                      </h3>
                      <p class="text-base md:text-lg text-taupe dark:text-text-muted-dark leading-relaxed md:ml-auto md:max-w-md">
                        {t(`about.principles.${key}.body`)}
                      </p>
                    </div>
                    <div class="md:col-start-9 md:col-span-2 md:row-start-1 order-1 md:order-2">
                      <span class="block text-7xl md:text-9xl font-black text-primary tracking-tighter leading-none select-none">
                        {num}
                      </span>
                    </div>
                    <div class="md:col-start-1 md:col-span-9 md:row-start-2 order-3">
                      {mockBlock}
                    </div>
                  </Reveal>
                );
              }
              // even (feed / analytics): num left, text middle, mock below
              return (
                <Reveal
                  key={key}
                  class="md:grid md:grid-cols-12 md:items-start md:gap-y-8 flex flex-col gap-8"
                >
                  <div class="md:col-start-1 md:col-span-2 md:row-start-1 order-1">
                    <span class="block text-7xl md:text-9xl font-black text-primary tracking-tighter leading-none select-none">
                      {num}
                    </span>
                  </div>
                  <div
                    class={`md:col-span-6 md:row-start-1 space-y-4 md:space-y-6 order-2 ${
                      idx === 0 ? "md:col-start-4" : "md:col-start-5"
                    }`}
                  >
                    <h3 class="text-2xl md:text-4xl font-bold text-charcoal dark:text-white leading-tight">
                      {t(`about.principles.${key}.title`)}
                    </h3>
                    <p class="text-base md:text-lg text-taupe dark:text-text-muted-dark leading-relaxed md:max-w-md">
                      {t(`about.principles.${key}.body`)}
                    </p>
                  </div>
                  <div class="md:col-start-4 md:col-span-9 md:row-start-2 order-3">
                    {mockBlock}
                  </div>
                </Reveal>
              );
            })}
          </div>
        </section>

        {/* Final CTA */}
        <section class="bg-card-light dark:bg-card-dark px-6 md:px-8 py-14 md:py-20">
          <div class="text-center max-w-5xl mx-auto">
          <div class="space-y-8 md:space-y-10">
            <p class="whitespace-pre-line text-xl md:text-3xl font-bold text-charcoal dark:text-white max-w-2xl mx-auto leading-snug">
              <Trans
                i18nKey="about.finalCta.description"
                components={{ b: <strong class="font-extrabold text-primary" /> }}
              />
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
                  class="underline hover:text-charcoal dark:hover:text-white"
                >
                  {t("common.termsLink")}
                </a>
                {t("common.privacyPolicyConsentMiddle")}
                <a
                  href="/privacy"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="underline hover:text-charcoal dark:hover:text-white"
                >
                  {t("common.privacyPolicyLink")}
                </a>
                {t("common.privacyPolicyConsentAfter")}
              </p>
            </div>
          </div>
          </div>
        </section>

        {/* Footer */}
        <footer class="border-t border-border-light dark:border-border-dark">
          <div class="max-w-7xl mx-auto px-6 md:px-8 py-6 flex flex-col md:flex-row justify-between items-center gap-4 text-sm text-taupe dark:text-text-muted-dark">
            <div class="flex items-center gap-6">
              <a
                href="/terms"
                class="font-medium no-underline text-inherit hover:text-primary"
              >
                {t("legal.navTerms")}
              </a>
              <a
                href="/privacy"
                class="font-medium no-underline text-inherit hover:text-primary"
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
