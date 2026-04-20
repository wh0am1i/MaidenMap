import { useTranslation } from "react-i18next";

export default function About() {
  const { t } = useTranslation();
  return (
    <div className="max-w-3xl mx-auto p-6 prose prose-invert prose-sm">
      <h1>{t("about.title")}</h1>
      <p>{t("about.intro")}</p>

      <h2>{t("about.api_heading")}</h2>
      <pre className="bg-[rgb(var(--panel-2))] p-3 rounded text-xs overflow-x-auto">
{`GET /api/grid/:code
GET /api/grid?codes=A,B,C
GET /api/health

# example
curl https://example.com/api/grid/JO65ab`}
      </pre>

      <h2>{t("about.credits_heading")}</h2>
      <p>{t("about.credits_body")}</p>
    </div>
  );
}
