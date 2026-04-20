import { useTranslation } from "react-i18next";
import { Textarea } from "@/components/ui/textarea";

export function BatchInput({
  value,
  onChange,
}: {
  value: string;
  onChange: (v: string) => void;
}) {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col gap-1">
      <label className="text-xs uppercase tracking-wider text-[rgb(var(--dim))]">
        {t("batch.paste_hint")}
      </label>
      <Textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        rows={5}
        className="font-mono text-sm"
      />
    </div>
  );
}

export function parseCodes(raw: string): string[] {
  return raw
    .split(/[,\n\s]+/)
    .map((s) => s.trim())
    .filter((s) => s.length > 0);
}
