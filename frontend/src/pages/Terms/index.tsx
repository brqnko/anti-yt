import { useTranslation } from "react-i18next";
import { LegalLayout } from "../../components/LegalLayout";
import { LegalSections, type LegalSection } from "../../components/LegalSections";
import { useMeta } from "../../hooks/useMeta";

const sections: LegalSection[] = [
  { num: "01", key: "application" },
  { num: "02", key: "registration" },
  { num: "03", key: "account" },
  { num: "04", key: "prohibited", items: true },
  { num: "05", key: "suspension" },
  { num: "06", key: "restriction" },
  { num: "07", key: "withdrawal" },
  { num: "08", key: "disclaimer" },
  { num: "09", key: "serviceChange" },
  { num: "10", key: "termsChange" },
  { num: "11", key: "personalInfo" },
  { num: "12", key: "notification" },
  { num: "13", key: "assignment" },
  { num: "14", key: "governingLaw" },
];

export default function Terms() {
  const { t } = useTranslation();
  useMeta({
    title: t("terms.title"),
    description: t("terms.metaDescription"),
    canonicalPath: "/terms",
  });

  return (
    <LegalLayout>
      <LegalSections ns="terms" sections={sections} />
    </LegalLayout>
  );
}
