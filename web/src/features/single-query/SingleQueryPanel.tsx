import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { GridInput } from "./GridInput";
import { ResultCard } from "./ResultCard";
import { History, type HistoryItem } from "./History";
import { useGridQuery } from "@/hooks/useGridQuery";
import { useHistory } from "@/hooks/useHistory";
import type { BiName, GridResponse } from "@/types/api";

function bestLabel(d: GridResponse): BiName {
  if (d.city.en || d.city.zh) return d.city;
  if (d.admin1.en || d.admin1.zh) return d.admin1;
  if (d.country) return d.country.name;
  return { en: "", zh: "" };
}

export function SingleQueryPanel({
  code,
  onCodeChange,
}: {
  code: string;
  onCodeChange: (v: string) => void;
}) {
  const { t } = useTranslation();
  const query = useGridQuery(code);
  const history = useHistory();

  useEffect(() => {
    if (query.isSuccess && query.data) {
      const d = query.data;
      history.push({
        grid: d.grid,
        label: bestLabel(d),
        countryCode: d.country?.code ?? "",
        at: Date.now(),
      });
    }
    // history.push is stable via useCallback
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query.isSuccess, query.data]);

  const items: HistoryItem[] = history.entries;
  const loading = code.length > 0 && query.isFetching && !query.data;

  return (
    <div className="flex flex-col gap-3 h-full overflow-y-auto p-4">
      <GridInput value={code} onChange={onCodeChange} onSubmit={onCodeChange} autoFocus />
      {loading && <ResultSkeleton />}
      {query.isSuccess && query.data && <ResultCard data={query.data} />}
      {query.isError && (
        <p className="text-sm text-[rgb(var(--danger))]">{String(query.error?.message ?? "error")}</p>
      )}
      <details open>
        <summary className="cursor-pointer text-xs uppercase tracking-wider text-[rgb(var(--dim))] md:hidden">
          {t("history.mobile_summary", { n: history.entries.length })}
        </summary>
        <div className="mt-2 md:mt-0">
          <History items={items} onPick={onCodeChange} onRemove={history.remove} onClear={history.clear} />
        </div>
      </details>
    </div>
  );
}

function ResultSkeleton() {
  return (
    <div className="bg-[rgb(var(--panel-2))] border border-border rounded-lg overflow-hidden animate-pulse">
      <div className="h-8 border-b border-border bg-black/5" />
      <div className="p-4 space-y-2">
        <div className="h-3 w-3/4 bg-white/10 rounded" />
        <div className="h-3 w-2/3 bg-white/10 rounded" />
        <div className="h-3 w-4/5 bg-white/10 rounded" />
        <div className="h-3 w-1/2 bg-white/10 rounded" />
      </div>
    </div>
  );
}
