import { useTranslation } from "react-i18next";
import { cn } from "@/lib/utils";

export interface BatchOkRow {
  kind: "ok";
  grid: string;
  seq: number;
  country: { code: string; nameEn: string };
  label: string;
}
export interface BatchErrorRow {
  kind: "error";
  grid: string;
  message: string;
}
export type BatchRow = BatchOkRow | BatchErrorRow;

export function BatchTable({
  rows,
  activeGrid,
  onRowClick,
}: {
  rows: BatchRow[];
  activeGrid: string | null;
  onRowClick: (grid: string) => void;
}) {
  const { t } = useTranslation();
  return (
    <div className="bg-[rgb(var(--panel-2))] border border-border rounded-md overflow-auto">
      <table className="w-full text-xs">
        <thead className="text-[10px] uppercase tracking-wider text-[rgb(var(--dim))]">
          <tr>
            <th className="text-left px-3 py-2 w-[70px]">CODE</th>
            <th className="w-[24px]"></th>
            <th className="text-left px-3 py-2">LOCATION</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => {
            const active = r.grid === activeGrid;
            return (
              <tr
                key={r.grid}
                data-active={active}
                onClick={() => onRowClick(r.grid)}
                className={cn(
                  "border-t border-border cursor-pointer",
                  active && "bg-[rgb(var(--ham))]/10",
                  r.kind === "error" && "opacity-50",
                )}
              >
                <td className="px-3 py-2 font-mono text-[rgb(var(--ham))]">{r.grid}</td>
                <td className="text-center">{r.kind === "ok" ? (
                  <span className="font-mono text-xs">{countryToFlag(r.country.code)}</span>
                ) : (
                  <span>⚠</span>
                )}</td>
                <td className="px-3 py-2 text-[rgb(var(--dim))]">
                  {r.kind === "ok" ? `${r.label}, ${r.country.nameEn}` : t("batch.invalid_format")}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function countryToFlag(code: string): string {
  if (code.length !== 2) return "·";
  const A = 0x1f1e6;
  const codePoints = [...code.toUpperCase()].map((c) => A + (c.charCodeAt(0) - 65));
  return String.fromCodePoint(...codePoints);
}
