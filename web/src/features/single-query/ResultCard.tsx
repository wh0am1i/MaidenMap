import { useTranslation } from "react-i18next";
import { Copy, Share2, Download } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { LocalizedName } from "@/components/LocalizedName";
import type { BiName, GridResponse } from "@/types/api";

export function ResultCard({ data }: { data: GridResponse }) {
  const { t } = useTranslation();

  async function copy() {
    await navigator.clipboard.writeText(buildSummary(data));
    toast.success(t("action.result_copied"));
  }

  async function share() {
    const url = new URL(window.location.href);
    url.searchParams.set("grid", data.grid);
    await navigator.clipboard.writeText(url.toString());
    toast.success(t("action.link_copied"));
  }

  return (
    <div className="bg-[rgb(var(--panel-2))] border border-border rounded-lg overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 border-b border-border text-xs uppercase tracking-wider text-[rgb(var(--dim))]">
        <span>{t("field.center")}</span>
        <span className="text-[rgb(var(--ham))] font-mono font-semibold">{data.grid}</span>
      </div>
      <dl className="px-4 py-3 text-sm">
        <Row label={t("field.country")}>
          {data.country ? (
            <>
              <LocalizedName value={data.country.name} />
              <span className="ml-2 font-mono text-xs text-[rgb(var(--dim))]">· {data.country.code}</span>
            </>
          ) : (
            "—"
          )}
        </Row>
        <Row label={t("field.admin1")}>
          <LocalizedName value={data.admin1} />
        </Row>
        <Row label={t("field.admin2")}>
          <LocalizedName value={data.admin2} />
        </Row>
        <Row label={t("field.city")}>
          <LocalizedName value={data.city} />
        </Row>
        <Row label={t("field.center")}>
          <span className="font-mono">
            {fmtLat(data.center.lat)} · {fmtLon(data.center.lon)}
          </span>
        </Row>
      </dl>
      <div className="flex gap-2 px-4 py-2 border-t border-border bg-black/10">
        <Button size="sm" variant="outline" onClick={copy}>
          <Copy size={12} /> {t("action.copy")}
        </Button>
        <Button size="sm" variant="outline" onClick={share}>
          <Share2 size={12} /> {t("action.share")}
        </Button>
        <Button size="sm" variant="outline" className="ml-auto" disabled>
          <Download size={12} /> {t("action.export")}
        </Button>
      </div>
    </div>
  );
}

function fmtLat(lat: number): string {
  return `${Math.abs(lat).toFixed(4).replace(/\.?0+$/, "")}°${lat >= 0 ? "N" : "S"}`;
}

function fmtLon(lon: number): string {
  return `${Math.abs(lon).toFixed(4).replace(/\.?0+$/, "")}°${lon >= 0 ? "E" : "W"}`;
}

// Bilingual summary — always includes both EN + ZH when both are present,
// so the copied text is useful regardless of the reader's language.
function buildSummary(d: GridResponse): string {
  const line = (k: string, v: BiName): string => {
    const en = v.en.trim();
    const zh = v.zh.trim();
    if (!en && !zh) return `${k}: —`;
    if (en && zh && en !== zh) return `${k}: ${en} / ${zh}`;
    return `${k}: ${en || zh}`;
  };
  const lines = [`Grid: ${d.grid}`];
  if (d.country) {
    const en = d.country.name.en;
    const zh = d.country.name.zh;
    const name = en && zh && en !== zh ? `${en} / ${zh}` : en || zh || "—";
    lines.push(`Country: ${name} (${d.country.code})`);
  } else {
    lines.push(`Country: —`);
  }
  lines.push(line("Admin1", d.admin1));
  lines.push(line("Admin2", d.admin2));
  lines.push(line("City", d.city));
  lines.push(`Center: ${fmtLat(d.center.lat)}, ${fmtLon(d.center.lon)}`);
  return lines.join("\n");
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex justify-between items-baseline py-[6px] border-b border-border last:border-none">
      <dt className="text-xs uppercase tracking-wider text-[rgb(var(--dim))]">{label}</dt>
      <dd className="text-sm text-right">{children}</dd>
    </div>
  );
}
