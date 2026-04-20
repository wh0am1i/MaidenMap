import { describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/mocks/server";
import { ApiError, getGrid, getGridBatch, getHealth } from "./api";

describe("api", () => {
  it("getGrid returns a single response", async () => {
    server.use(
      http.get("/api/grid/JO65ab", () =>
        HttpResponse.json({
          grid: "JO65ab",
          center: { lat: 55.0625, lon: 12.0417 },
          country: { code: "DK", name: { en: "Denmark", zh: "丹麦" } },
          admin1: { en: "Zealand", zh: "西兰" },
          admin2: { en: "", zh: "" },
          city: { en: "Vordingborg", zh: "" },
        }),
      ),
    );
    const r = await getGrid("JO65ab");
    expect(r.grid).toBe("JO65ab");
    expect(r.country?.code).toBe("DK");
    expect(r.country?.name.zh).toBe("丹麦");
  });

  it("getGridBatch returns a results array", async () => {
    server.use(
      http.get("/api/grid", ({ request }) => {
        const codes = new URL(request.url).searchParams.get("codes");
        expect(codes).toBe("JO65ab,BAD");
        return HttpResponse.json({
          results: [
            {
              grid: "JO65ab",
              center: { lat: 55.06, lon: 12.04 },
              country: { code: "DK", name: { en: "Denmark", zh: "丹麦" } },
              admin1: { en: "", zh: "" },
              admin2: { en: "", zh: "" },
              city: { en: "", zh: "" },
            },
            { grid: "BAD", error: "invalid_grid", message: "bad" },
          ],
        });
      }),
    );
    const r = await getGridBatch(["JO65ab", "BAD"]);
    expect(r.results).toHaveLength(2);
  });

  it("throws ApiError on 400 invalid_grid", async () => {
    server.use(
      http.get("/api/grid/BAD", () =>
        HttpResponse.json(
          { error: "invalid_grid", message: "invalid length" },
          { status: 400 },
        ),
      ),
    );
    await expect(getGrid("BAD")).rejects.toMatchObject({
      name: "ApiError",
      status: 400,
      code: "invalid_grid",
    });
  });

  it("throws ApiError with code rate_limited on 429", async () => {
    server.use(
      http.get("/api/grid/JO65", () =>
        HttpResponse.json(
          { error: "rate_limited", message: "slow down" },
          { status: 429 },
        ),
      ),
    );
    await expect(getGrid("JO65")).rejects.toMatchObject({
      name: "ApiError",
      status: 429,
      code: "rate_limited",
    });
  });

  it("getHealth returns counts", async () => {
    const r = await getHealth();
    expect(r.status).toBe("ok");
    expect(r.cities_count).toBeGreaterThan(0);
  });

  it("ApiError is an Error subclass", () => {
    const e = new ApiError(500, "network", "boom");
    expect(e).toBeInstanceOf(Error);
    expect(e.name).toBe("ApiError");
  });
});
