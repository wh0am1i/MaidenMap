import { QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it } from "vitest";
import { server } from "@/mocks/server";
import { createQueryClient } from "@/lib/query-client";
import { qk } from "@/lib/query-keys";
import { useGridQuery, useGridBatchQuery } from "./useGridQuery";

function wrap() {
  const qc = createQueryClient();
  const Wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={qc}>{children}</QueryClientProvider>
  );
  return { qc, Wrapper };
}

const sampleResp = (grid: string, city: string) => ({
  grid,
  center: { lat: 55, lon: 12 },
  country: { code: "DK", name: { en: "Denmark", zh: "丹麦" } },
  admin1: { en: "", zh: "" },
  admin2: { en: "", zh: "" },
  city: { en: city, zh: "" },
});

describe("useGridQuery (single)", () => {
  beforeEach(() => server.resetHandlers());

  it("fetches and caches a single code", async () => {
    server.use(http.get("/api/grid/JO65", () => HttpResponse.json(sampleResp("JO65", "Malmö"))));
    const { qc, Wrapper } = wrap();
    const { result } = renderHook(() => useGridQuery("JO65"), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.city.en).toBe("Malmö");
    expect(qc.getQueryData(qk.grid("JO65"))).toBeDefined();
  });

  it("disabled when code is empty", () => {
    const { Wrapper } = wrap();
    const { result } = renderHook(() => useGridQuery(""), { wrapper: Wrapper });
    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useGridBatchQuery", () => {
  beforeEach(() => server.resetHandlers());

  it("seeds per-code cache from batch results", async () => {
    server.use(
      http.get("/api/grid", ({ request }) => {
        const codes = new URL(request.url).searchParams.get("codes")!;
        return HttpResponse.json({
          results: codes.split(",").map((g) => sampleResp(g, `city-${g}`)),
        });
      }),
    );
    const { qc, Wrapper } = wrap();
    const { result } = renderHook(() => useGridBatchQuery(["JO65", "PM95"]), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const stored = qc.getQueryData(qk.grid("JO65"));
    expect(stored).toBeDefined();
    expect((stored as { city: { en: string } }).city.en).toBe("city-JO65");
  });

  it("short-circuits codes already in cache", async () => {
    const { qc, Wrapper } = wrap();
    qc.setQueryData(qk.grid("JO65"), sampleResp("JO65", "cached"));

    let netHit = 0;
    server.use(
      http.get("/api/grid", ({ request }) => {
        netHit++;
        const codes = new URL(request.url).searchParams.get("codes")!;
        return HttpResponse.json({
          results: codes.split(",").map((g) => sampleResp(g, `net-${g}`)),
        });
      }),
    );
    const { result } = renderHook(() => useGridBatchQuery(["JO65", "PM95"]), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    // JO65 was cached; only PM95 should have been requested
    expect(netHit).toBe(1);
    const stored = qc.getQueryData(qk.grid("JO65"));
    expect((stored as { city: { en: string } }).city.en).toBe("cached");
  });
});
