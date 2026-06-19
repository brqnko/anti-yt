import { useTranslation } from "react-i18next";
import { LegalLayout } from "../../components/LegalLayout";
import { LegalSections, type LegalSection } from "../../components/LegalSections";
import { useMeta } from "../../hooks/useMeta";

const sections: LegalSection[] = [
  { num: "01", key: "definition" },
  { num: "02", key: "collection" },
  { num: "03", key: "purpose", items: true },
  { num: "04", key: "purposeChange" },
  { num: "05", key: "thirdParty", items: true },
  { num: "06", key: "disclosure" },
  { num: "07", key: "correction" },
  { num: "08", key: "suspensionOfUse" },
  { num: "09", key: "policyChange" },
  {
    num: "10",
    key: "contact",
    link: {
      url: "mailto:contact@brqnko.rs",
      labelKey: "privacy.sections.contact.linkLabel",
    },
  },
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
      <LegalSections ns="privacy" sections={sections} />
    </LegalLayout>
  );
}
