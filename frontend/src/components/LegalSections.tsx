import { useTranslation } from "react-i18next";

export type LegalSection = {
  num: string;
  key: string;
  items?: boolean;
  link?: { url: string; labelKey: string };
};

export function LegalSections({
  ns,
  sections,
}: {
  ns: "terms" | "privacy";
  sections: LegalSection[];
}) {
  const { t } = useTranslation();

  return (
    <>
      <div class="mb-16 text-center">
        <h1 class="text-4xl md:text-6xl font-black leading-tight tracking-tight mb-6">
          {t(`${ns}.title`)}
        </h1>
      </div>

      <div class="max-w-4xl mx-auto">
        <div class="bg-white dark:bg-[#26221c] border border-[#e1ddd6] dark:border-[#3d372e] rounded-2xl overflow-hidden mb-12">
          <div class="p-10 md:p-16 text-lg text-[#3a352e] dark:text-[#d1cdc7] leading-8 tracking-[0.01em]">
            <p class="mt-0 mb-12 whitespace-pre-line">{t(`${ns}.intro`)}</p>

            {sections.map((s) => {
              const items = s.items
                ? (t(`${ns}.sections.${s.key}.items`, {
                    returnObjects: true,
                  }) as string[])
                : null;

              return (
                <section class="mb-12" key={s.key}>
                  <h4 class="font-bold text-xl mb-6 text-[#171511] dark:text-white flex items-center gap-3 m-0">
                    <span class="text-primary">{s.num}.</span>{" "}
                    {t(`${ns}.sections.${s.key}.title`)}
                  </h4>
                  <p class="m-0 whitespace-pre-line">
                    {t(`${ns}.sections.${s.key}.body`)}
                  </p>
                  {Array.isArray(items) && (
                    <ol class="list-decimal pl-6 mt-4 space-y-2">
                      {items.map((item, i) => (
                        <li key={i}>{item}</li>
                      ))}
                    </ol>
                  )}
                  {s.link && (
                    <p class="m-0 mt-4">
                      <a
                        href={s.link.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="text-primary font-bold no-underline hover:underline"
                      >
                        {t(s.link.labelKey)}
                      </a>
                    </p>
                  )}
                </section>
              );
            })}

            <p class="m-0 mt-16 text-base text-[#847862] dark:text-[#a89d89] text-right">
              {t(`${ns}.effectiveDate`)}
            </p>
          </div>
        </div>
      </div>
    </>
  );
}
