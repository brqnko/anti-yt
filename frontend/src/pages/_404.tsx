import { useTranslation } from "react-i18next";
import { useTitle } from "../hooks/useTitle";
export default function NotFound() {
  const { t } = useTranslation();
  useTitle("404 Not Found");

  return (
    <main class="flex flex-1 flex-col items-center justify-center px-6 text-center max-w-4xl mx-auto w-full py-24">
      <div class="flex flex-col gap-4 mb-10 mt-8">
        <h1 class="text-charcoal dark:text-slate-100 text-6xl md:text-8xl leading-tight tracking-tighter font-bold">
          404
        </h1>
        <p class="text-taupe dark:text-slate-400 text-base md:text-lg max-w-md mx-auto leading-relaxed">
          {t("common.notFoundMessage")}
        </p>
      </div>
      <div class="flex justify-center w-full">
        <a
          href="/"
          class="flex items-center justify-center rounded-xl h-14 px-10 bg-primary text-white text-base font-bold transition-transform hover:scale-[1.02] active:scale-[0.98] min-w-[240px] no-underline"
        >
          {t("common.returnHome")}
        </a>
      </div>
    </main>
  );
}
