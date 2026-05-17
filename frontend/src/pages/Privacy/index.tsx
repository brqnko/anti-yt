import { useTranslation } from "react-i18next";
import { LegalLayout } from "../../components/LegalLayout";
import { useMeta } from "../../hooks/useMeta";

const sections = [
  { num: "01", key: "dataCollection" },
  { num: "02", key: "dataUsage" },
  { num: "03", key: "dataSharing" },
  { num: "04", key: "cookies" },
  { num: "05", key: "security" },
  { num: "06", key: "userRights" },
  { num: "07", key: "changes" },
];

export default function Privacy() {
  const { t } = useTranslation();
  useMeta({
    title: t("privacy.title"),
    description: t("privacy.metaDescription"),
    canonicalPath: "/privacy",
  });

  return (
    <LegalLayout>
      <div class="mb-16 text-center">
        <h1 class="text-4xl md:text-6xl font-black leading-tight tracking-tight mb-6">
          {t("privacy.title")}
        </h1>
      </div>

      <div class="max-w-4xl mx-auto">
        <div class="bg-white dark:bg-[#26221c] border border-[#e1ddd6] dark:border-[#3d372e] rounded-2xl overflow-hidden mb-12">

          <div class="p-10 md:p-16 text-lg text-[#3a352e] dark:text-[#d1cdc7] leading-8 tracking-[0.01em]">
            {sections.map((s) => (
              <section class="mb-12" key={s.key}>
                <h4 class="font-bold text-xl mb-6 text-[#171511] dark:text-white flex items-center gap-3 m-0">
                  <span class="text-primary">{s.num}.</span>{" "}
                  {t(`privacy.sections.${s.key}.title`)}
                </h4>
                <p class="m-0">{t(`privacy.sections.${s.key}.body`)}</p>
              </section>
            ))}
          </div>
        </div>
      </div>
    </LegalLayout>
  );
}
