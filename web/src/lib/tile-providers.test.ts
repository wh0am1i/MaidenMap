import { describe, expect, it } from "vitest";
import { builtInProviders, validateCustomTileUrl } from "./tile-providers";

describe("tile-providers", () => {
  it("exposes Carto light + dark built-ins", () => {
    const ids = builtInProviders.map((p) => p.id);
    expect(ids).toContain("carto-light");
    expect(ids).toContain("carto-dark");
  });

  it.each([
    "https://example.com/tiles/{z}/{x}/{y}.png",
    "https://a.tile.osm.org/{z}/{x}/{y}.png",
  ])("accepts valid URL with z/x/y placeholders: %s", (url) => {
    expect(validateCustomTileUrl(url)).toEqual({ ok: true });
  });

  it("rejects URL without z/x/y placeholders", () => {
    const r = validateCustomTileUrl("https://example.com/no-placeholders.png");
    expect(r.ok).toBe(false);
    if (!r.ok) expect(r.reason).toContain("{z}");
  });

  it("rejects non-https URL", () => {
    const r = validateCustomTileUrl("ftp://foo/{z}/{x}/{y}.png");
    expect(r.ok).toBe(false);
  });
});
