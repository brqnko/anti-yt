import { useTranslation } from "react-i18next";

export default function NotFound() {
  const { t } = useTranslation();

  return (
    <section>
      <h1>{t("common.notFound")}</h1>
      <p>{t("common.notFoundMessage")}</p>
    </section>
  );
}
