export interface TileProvider {
  id: string;
  name: string;
  url: string;
  attribution: string;
  /** "light" | "dark" | "any" — used to auto-pick on theme change */
  variant: "light" | "dark" | "any";
  maxZoom: number;
}

export const builtInProviders: TileProvider[] = [
  {
    id: "carto-light",
    name: "Carto Positron",
    url: "https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png",
    attribution: "© OpenStreetMap contributors © CARTO",
    variant: "light",
    maxZoom: 19,
  },
  {
    id: "carto-dark",
    name: "Carto Dark Matter",
    url: "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png",
    attribution: "© OpenStreetMap contributors © CARTO",
    variant: "dark",
    maxZoom: 19,
  },
  {
    id: "osm",
    name: "OpenStreetMap",
    url: "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
    attribution: "© OpenStreetMap contributors",
    variant: "any",
    maxZoom: 19,
  },
];

export type ValidationResult = { ok: true } | { ok: false; reason: string };

export function validateCustomTileUrl(url: string): ValidationResult {
  if (!/^https:\/\//.test(url)) {
    return { ok: false, reason: "URL must start with https://" };
  }
  const hasZ = url.includes("{z}");
  const hasX = url.includes("{x}");
  const hasY = url.includes("{y}");
  if (!(hasZ && hasX && hasY)) {
    return { ok: false, reason: "URL must contain {z}, {x}, {y} placeholders" };
  }
  return { ok: true };
}
