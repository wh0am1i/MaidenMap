import { useSearchParams } from "react-router-dom";
import { useMemo, useState } from "react";
import { SingleQueryPanel } from "@/features/single-query/SingleQueryPanel";
import { BatchQueryPanel } from "@/features/batch-query/BatchQueryPanel";
import { MapView } from "@/features/map/MapView";
import { GridOverlay } from "@/features/map/GridOverlay";
import { BatchOverlay, type BatchPin } from "@/features/map/BatchOverlay";
import { useMapTile } from "@/hooks/useMapTile";
import { useTheme } from "@/hooks/useTheme";
import { useGridBatchQuery } from "@/hooks/useGridQuery";
import { parseMaidenhead } from "@/lib/maidenhead";
import { isGridError } from "@/types/api";

export default function Home() {
  const [params, setParams] = useSearchParams();
  const tab = params.get("tab") === "batch" ? "batch" : "single";
  const code = params.get("grid") ?? "";
  const { resolved } = useTheme();
  const { provider } = useMapTile(resolved);

  function setCode(v: string) {
    if (!v) params.delete("grid");
    else params.set("grid", v);
    setParams(params, { replace: true });
  }

  const [activeBatchGrid, setActiveBatchGrid] = useState<string | null>(null);

  const parsed = code ? parseMaidenhead(code) : null;
  const parsedOk = parsed && parsed.kind === "ok" ? parsed : null;

  // Batch pins (from cache or live query) — we rely on the panel having submitted codes.
  const batchSubmitted = useMemo(() => {
    if (tab !== "batch") return [];
    return (params.get("codes") ?? "").split(",").map((s) => s.trim()).filter(Boolean);
  }, [tab, params]);
  const batchQuery = useGridBatchQuery(batchSubmitted);
  const pins: BatchPin[] = useMemo(() => {
    if (!batchQuery.data) return [];
    let seq = 0;
    const out: BatchPin[] = [];
    for (const r of batchQuery.data.results) {
      if (isGridError(r)) continue;
      seq++;
      out.push({ grid: r.grid, seq, lat: r.center.lat, lon: r.center.lon });
    }
    return out;
  }, [batchQuery.data]);

  const defaultCenter: [number, number] = parsedOk ? [parsedOk.lat, parsedOk.lon] : [30, 10];
  const defaultZoom = parsedOk ? (parsedOk.length === 8 ? 12 : parsedOk.length === 6 ? 8 : 4) : 2;

  return (
    <div className="flex flex-col md:grid md:grid-cols-[360px_1fr] h-full min-h-0">
      <div className="order-2 md:order-1 md:h-full border-t md:border-t-0 md:border-r border-border bg-[rgb(var(--panel))] overflow-hidden min-h-0">
        {tab === "single" ? (
          <SingleQueryPanel code={code} onCodeChange={setCode} />
        ) : (
          <BatchQueryPanel activeGrid={activeBatchGrid} onActiveChange={setActiveBatchGrid} />
        )}
      </div>
      <div className="order-1 md:order-2 h-[45vh] md:h-full">
        <MapView provider={provider} center={defaultCenter} zoom={defaultZoom}>
          {tab === "single" && parsedOk && <GridOverlay grid={parsedOk} fitOnMount />}
          {tab === "batch" && pins.length > 0 && (
            <BatchOverlay pins={pins} activeGrid={activeBatchGrid} fitOnChange />
          )}
        </MapView>
      </div>
    </div>
  );
}
