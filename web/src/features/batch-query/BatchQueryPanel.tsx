import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { BatchInput, parseCodes } from "./BatchInput";
import { BatchTable, type BatchRow } from "./BatchTable";
import { useGridBatchQuery } from "@/hooks/useGridQuery";
import { parseMaidenhead } from "@/lib/maidenhead";
import { isGridError, type GridResponse } from "@/types/api";
import { toCSV, downloadCSV } from "@/lib/csv";

export function BatchQueryPanel({
  activeGrid,
  onActiveChange,
}: {
  activeGrid: string | null;
  onActiveChange: (grid: string | null) => void;
}) {
  const { t } = useTranslation();
  const [raw, setRaw] = useState("");
  const [submitted, setSubmitted] = useState<string[]>([]);

  const codes = parseCodes(raw);

  // client-side bad codes never hit the API; surface them as rows
  const { client, server } = useMemo(() => {
    const client: string[] = [];
    const server: string[] = [];
    for (const c of codes) {
      if (parseMaidenhead(c).kind === "ok") server.push(c);
      else client.push(c);
    }
    return { client, server };
  }, [codes]);

  const q = useGridBatchQuery(submitted.length > 0 ? submitted : []);

  const rows: BatchRow[] = useMemo(() => {
    const out: BatchRow[] = [];
    const serverMap = new Map(
      q.data?.results.map((r) => [("grid" in r ? r.grid : "") as string, r]) ?? [],
    );
    let seq = 0;
    for (const c of codes) {
      if (client.includes(c)) {
        out.push({ kind: "error", grid: c, message: t("batch.invalid_format") });
        continue;
      }
      const r = serverMap.get(c);
      if (!r) continue;
      if (isGridError(r)) {
        out.push({ kind: "error", grid: r.grid, message: r.message ?? r.error });
      } else {
        seq++;
        out.push({
          kind: "ok",
          grid: r.grid,
          seq,
          country: { code: r.country?.code ?? "", nameEn: r.country?.name.en ?? t("batch.no_country") },
          label: r.city.en || r.admin1.en || "",
        });
      }
    }
    return out;
  }, [codes, client, q.data, t]);

  const okCount = rows.filter((r) => r.kind === "ok").length;

  function exportCsv() {
    const headers = [
      "grid",
      "country_code",
      "country_name_en",
      "country_name_zh",
      "admin1_en",
      "admin1_zh",
      "admin2_en",
      "admin2_zh",
      "city_en",
      "city_zh",
      "lat",
      "lon",
    ];
    const data =
      q.data?.results
        .filter((r): r is GridResponse => !isGridError(r))
        .map((r) => [
          r.grid,
          r.country?.code ?? "",
          r.country?.name.en ?? "",
          r.country?.name.zh ?? "",
          r.admin1.en,
          r.admin1.zh,
          r.admin2.en,
          r.admin2.zh,
          r.city.en,
          r.city.zh,
          String(r.center.lat),
          String(r.center.lon),
        ]) ?? [];
    downloadCSV("maidenmap-batch.csv", toCSV(headers, data));
  }

  return (
    <div className="flex flex-col gap-3 h-full overflow-y-auto p-4">
      <BatchInput value={raw} onChange={setRaw} />
      <div className="flex gap-2">
        <Button onClick={() => setSubmitted(server)} disabled={server.length === 0}>
          ▶ {t("action.query_all", { n: server.length })}
        </Button>
        <Button variant="outline" onClick={() => {
          setRaw("");
          setSubmitted([]);
          onActiveChange(null);
        }}>
          {t("action.clear")}
        </Button>
      </div>
      <BatchTable rows={rows} activeGrid={activeGrid} onRowClick={onActiveChange} />
      <div className="flex items-center gap-2 text-xs text-[rgb(var(--dim))]">
        <span>{t("batch.results_summary", { ok: okCount, total: rows.length })}</span>
        <Button size="sm" variant="outline" className="ml-auto" onClick={exportCsv} disabled={okCount === 0}>
          {t("action.export_csv")}
        </Button>
      </div>
    </div>
  );
}
