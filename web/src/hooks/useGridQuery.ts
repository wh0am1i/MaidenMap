import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ApiError, getGrid, getGridBatch } from "@/lib/api";
import { qk } from "@/lib/query-keys";
import type { BatchResponse, GridResponse } from "@/types/api";

export function useGridQuery(code: string) {
  return useQuery<GridResponse, ApiError>({
    queryKey: qk.grid(code),
    queryFn: () => getGrid(code),
    enabled: code.length > 0,
  });
}

export function useGridBatchQuery(codes: string[]) {
  const qc = useQueryClient();
  return useQuery<BatchResponse, ApiError>({
    queryKey: qk.gridBatch(codes),
    enabled: codes.length > 0,
    queryFn: async () => {
      const missing = codes.filter((c) => qc.getQueryData(qk.grid(c)) === undefined);
      let netResults: BatchResponse["results"] = [];
      if (missing.length > 0) {
        const resp = await getGridBatch(missing);
        netResults = resp.results;
        // Seed the per-code cache for each success in the batch response
        for (const item of resp.results) {
          if (!("error" in item)) {
            qc.setQueryData(qk.grid(item.grid), item);
          }
        }
      }
      // Prefer the per-code cache for order stability with the original codes.
      const merged = codes.map((c) => {
        const cached = qc.getQueryData<GridResponse>(qk.grid(c));
        if (cached) return cached;
        const netHit = netResults.find((r) => r.grid === c);
        if (netHit) return netHit;
        return { grid: c, error: "not_found", message: "" };
      });
      return { results: merged };
    },
  });
}
