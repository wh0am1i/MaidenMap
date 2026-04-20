import { useTranslation } from "react-i18next";
import type { BiName } from "@/types/api";

export function LocalizedName({ value, className }: { value: BiName; className?: string }) {
  const { i18n } = useTranslation();
  const lang = i18n.language;
  let pick = "";
  if (lang.startsWith("zh")) pick = value.zh || value.en;
  else pick = value.en || value.zh;
  if (!pick) pick = "—";
  return <span className={className}>{pick}</span>;
}
