import { useTranslation } from "react-i18next";
import { languages } from "../../constants";
import { Icon } from "../../components/Icon";

interface Step1Props {
  displayName: string;
  setDisplayName: (v: string) => void;
  languageCode: string;
  setLanguageCode: (v: string) => void;
  onNext: () => void;
}

export default function Step1({ displayName, setDisplayName, languageCode, setLanguageCode, onNext }: Step1Props) {
  const { t } = useTranslation();

  const trimmedName = displayName.trim();
  const nameLength = [...trimmedName].length;
  const isNameValid = nameLength >= 3 && nameLength <= 29;

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    if (!isNameValid) return;
    onNext();
  };

  return (
    <>
      {/* Step header */}
      <div class="mb-8 text-center">
        <span class="inline-block py-1 px-3 rounded-full bg-primary/10 text-primary text-xs font-bold uppercase tracking-wider mb-3">
          {t("register.step", { current: 1, total: 2 })}
        </span>
        <h2 class="text-3xl font-black text-charcoal dark:text-white mb-2">
          {t("register.profileDetails.title")}
        </h2>
        <p class="text-taupe dark:text-gray-400">
          {t("register.profileDetails.subtitle")}
        </p>
      </div>

      {/* Card */}
      <div class="bg-white dark:bg-[#2a2721] rounded-2xl border border-gray-100 dark:border-neutral-800 p-8 md:p-10 relative overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-primary to-transparent opacity-50" />

        <form class="space-y-10" onSubmit={handleSubmit}>
          {/* Display Name */}
          <div class="space-y-4">
            <div class="flex items-center gap-3 mb-2">
              <div class="flex items-center justify-center size-8 rounded-full bg-background-light dark:bg-neutral-800 text-charcoal dark:text-white font-bold border border-gray-200 dark:border-neutral-700 text-sm">
                1
              </div>
              <label class="text-lg font-bold text-charcoal dark:text-white" for="display-name">
                {t("register.profileDetails.displayName")}
              </label>
            </div>
            <div class="ml-11">
              <input
                id="display-name"
                type="text"
                class="w-full px-4 py-3 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white placeholder-taupe focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all"
                placeholder={t("register.profileDetails.displayNamePlaceholder")}
                value={displayName}
                onInput={(e) => setDisplayName((e.target as HTMLInputElement).value)}
                required
              />
              <div class="mt-2 flex items-center justify-end">
                <span class={`text-xs tabular-nums ${nameLength > 29 ? "text-red-500" : "text-taupe"}`}>
                  {nameLength}/29
                </span>
              </div>
            </div>
          </div>

          <div class="h-px w-full bg-gray-100 dark:bg-neutral-800" />

          {/* Content Language */}
          <div class="space-y-4">
            <div class="flex items-center gap-3 mb-2">
              <div class="flex items-center justify-center size-8 rounded-full bg-background-light dark:bg-neutral-800 text-charcoal dark:text-white font-bold border border-gray-200 dark:border-neutral-700 text-sm">
                2
              </div>
              <label class="text-lg font-bold text-charcoal dark:text-white" for="language-select">
                {t("register.profileDetails.contentLanguage")}
              </label>
            </div>
            <div class="ml-11">
              <div class="relative">
                <Icon name="translate" class="absolute left-4 top-1/2 -translate-y-1/2   text-taupe" />
                <select
                  id="language-select"
                  class="w-full pl-12 pr-10 py-3 rounded-xl bg-background-light dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-charcoal dark:text-white focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none appearance-none transition-all cursor-pointer"
                  value={languageCode}
                  onChange={(e) => setLanguageCode((e.target as HTMLSelectElement).value)}
                >
                  {languages.map((lang) => (
                    <option key={lang.code} value={lang.code}>
                      {lang.label}
                    </option>
                  ))}
                </select>
                <Icon name="expand_more" class="absolute right-4 top-1/2 -translate-y-1/2   text-taupe pointer-events-none" />
              </div>
            </div>
          </div>

          {/* Submit */}
          <div class="pt-6 flex items-center justify-end gap-4">
            <button
              type="submit"
              disabled={!isNameValid}
              class="w-full sm:w-auto px-8 py-3.5 bg-primary hover:bg-primary/90 text-white font-bold rounded-xl transition-all flex items-center justify-center gap-2 group cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed border-none"
            >
              <span>{t("register.continue")}</span>
              <Icon name="arrow_forward" class="group-hover:translate-x-1 transition-transform" />
            </button>
          </div>
        </form>
      </div>
    </>
  );
}
