// Wire types mirroring backend `/api/grid` responses (see backend L3 plan Task 11).

export interface BiName {
  en: string;
  zh: string;
}

export interface GridCenter {
  lat: number;
  lon: number;
}

export interface CountryInfo {
  code: string;
  name: BiName;
}

export interface GridResponse {
  grid: string;
  center: GridCenter;
  country: CountryInfo | null;
  admin1: BiName;
  admin2: BiName;
  city: BiName;
}

export interface GridError {
  grid: string;
  error: string;
  message?: string;
}

export type BatchItem = GridResponse | GridError;

export interface BatchResponse {
  results: BatchItem[];
}

export interface HealthResponse {
  status: "ok";
  cities_count: number;
  countries_count: number;
  data_updated_at: string; // ISO 8601
}

export function isGridError(item: BatchItem): item is GridError {
  return (item as GridError).error !== undefined;
}
