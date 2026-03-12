import { useTranslation } from "react-i18next";
import { useTitle } from "../hooks/useTitle";

export default function NotFound() {
  const { t } = useTranslation();
  useTitle("404 Not Found");

  return (
    <section>
      <h1>{t("common.notFound")}</h1>
      <p>{t("common.notFoundMessage")}</p>
    </section>
  );
}
