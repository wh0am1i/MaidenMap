import { useEffect } from "react";
import { GridInput } from "./GridInput";
import { ResultCard } from "./ResultCard";
import { History, type HistoryItem } from "./History";
import { useGridQuery } from "@/hooks/useGridQuery";
import { useHistory } from "@/hooks/useHistory";

export function SingleQueryPanel({
  code,
  onCodeChange,
}: {
  code: string;
  onCodeChange: (v: string) => void;
}) {
  const query = useGridQuery(code);
  const history = useHistory();

  useEffect(() => {
    if (query.isSuccess && query.data) {
      const d = query.data;
      const label = d.city.en || d.admin1.en || d.country?.name.en || "";
      history.push({
        grid: d.grid,
        label,
        countryCode: d.country?.code ?? "",
        at: Date.now(),
      });
    }
    // history.push is stable via useCallback
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query.isSuccess, query.data]);

  const items: HistoryItem[] = history.entries;

  return (
    <div className="flex flex-col gap-3 h-full overflow-y-auto p-4">
      <GridInput value={code} onChange={onCodeChange} onSubmit={onCodeChange} autoFocus />
      {query.isSuccess && query.data && <ResultCard data={query.data} />}
      {query.isError && (
        <p className="text-sm text-[rgb(var(--danger))]">{String(query.error?.message ?? "error")}</p>
      )}
      <History items={items} onPick={onCodeChange} onClear={history.clear} />
    </div>
  );
}
