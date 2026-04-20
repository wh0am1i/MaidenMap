import { http, HttpResponse } from "msw";

// Default handlers — individual tests override as needed via server.use(...).
export const handlers = [
  http.get("/api/health", () =>
    HttpResponse.json({
      status: "ok",
      cities_count: 100,
      countries_count: 200,
      data_updated_at: "2026-04-20T00:00:00Z",
    }),
  ),
];
