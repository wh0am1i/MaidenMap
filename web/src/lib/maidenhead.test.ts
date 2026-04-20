import { describe, expect, it } from "vitest";
import { parseMaidenhead } from "./maidenhead";

describe("parseMaidenhead", () => {
  it.each<[string, number, number, number]>([
    ["JJ55", 5.5, 11.0, 4],
    ["JO65", 55.5, 13.0, 4],
    ["JO65ab", 55.0625, 12.0417, 6],
    ["PM95", 35.5, 139.0, 4],
    ["OM89", 39.5, 117.0, 4],
  ])("parses %s to (%f, %f) length %d", (code, lat, lon, length) => {
    const r = parseMaidenhead(code);
    expect(r.kind).toBe("ok");
    if (r.kind !== "ok") return;
    expect(r.lat).toBeCloseTo(lat, 3);
    expect(r.lon).toBeCloseTo(lon, 3);
    expect(r.length).toBe(length);
    expect(r.normalized).toBe(code);
  });

  it("uppercases field / lowercases subsquare on normalize", () => {
    const r = parseMaidenhead("jo65AB");
    expect(r.kind).toBe("ok");
    if (r.kind !== "ok") return;
    expect(r.normalized).toBe("JO65ab");
  });

  it("rejects wrong length", () => {
    const r = parseMaidenhead("JO6");
    expect(r.kind).toBe("error");
    if (r.kind !== "error") return;
    expect(r.reason).toBe("length");
  });

  it("rejects wrong field character", () => {
    const r = parseMaidenhead("ZZ00");
    expect(r.kind).toBe("error");
    if (r.kind !== "error") return;
    expect(r.reason).toBe("field");
  });

  it("rejects wrong digit", () => {
    const r = parseMaidenhead("JOab");
    expect(r.kind).toBe("error");
    if (r.kind !== "error") return;
    expect(r.reason).toBe("square");
  });

  it("rejects wrong subsquare character for 6-length", () => {
    const r = parseMaidenhead("JO6599");
    expect(r.kind).toBe("error");
    if (r.kind !== "error") return;
    expect(r.reason).toBe("subsquare");
  });

  it("gives bounds for a 6-char grid (bounds are inclusive-lo exclusive-hi)", () => {
    const r = parseMaidenhead("JO65ab");
    expect(r.kind).toBe("ok");
    if (r.kind !== "ok") return;
    // JO65ab lon bounds: 12° + a(0)*5' to +5' => 12°0' to 12°5' => 12.0 to 12.08333°
    expect(r.bounds.lonMin).toBeCloseTo(12.0, 3);
    expect(r.bounds.lonMax).toBeCloseTo(12.0833, 3);
    expect(r.bounds.latMin).toBeCloseTo(55.0417, 3);
    expect(r.bounds.latMax).toBeCloseTo(55.0833, 3);
  });
});
