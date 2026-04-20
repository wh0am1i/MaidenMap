import type { BatchResponse, GridResponse, HealthResponse } from "@/types/api";

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

async function request<T>(path: string): Promise<T> {
  let resp: Response;
  try {
    resp = await fetch(path, { headers: { Accept: "application/json" } });
  } catch (err) {
    throw new ApiError(0, "network", err instanceof Error ? err.message : String(err));
  }
  const text = await resp.text();
  let body: unknown = undefined;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      // non-JSON response
    }
  }
  if (!resp.ok) {
    const errBody = body as { error?: string; message?: string } | undefined;
    throw new ApiError(
      resp.status,
      errBody?.error ?? `http_${resp.status}`,
      errBody?.message ?? resp.statusText,
    );
  }
  return body as T;
}

export function getGrid(code: string): Promise<GridResponse> {
  return request<GridResponse>(`/api/grid/${encodeURIComponent(code)}`);
}

export function getGridBatch(codes: string[]): Promise<BatchResponse> {
  const q = new URLSearchParams({ codes: codes.join(",") });
  return request<BatchResponse>(`/api/grid?${q.toString()}`);
}

export function getHealth(): Promise<HealthResponse> {
  return request<HealthResponse>("/api/health");
}
