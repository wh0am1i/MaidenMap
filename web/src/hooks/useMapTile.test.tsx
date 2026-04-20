import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { useMapTile } from "./useMapTile";

describe("useMapTile", () => {
  beforeEach(() => localStorage.clear());

  it("defaults to carto-light for light theme", () => {
    const { result } = renderHook(() => useMapTile("light"));
    expect(result.current.provider.id).toBe("carto-light");
  });

  it("defaults to carto-dark for dark theme", () => {
    const { result } = renderHook(() => useMapTile("dark"));
    expect(result.current.provider.id).toBe("carto-dark");
  });

  it("user choice overrides theme default", () => {
    const { result } = renderHook(() => useMapTile("light"));
    act(() => result.current.setProviderId("osm"));
    expect(result.current.provider.id).toBe("osm");
  });

  it("falls back to theme default on invalid stored id", () => {
    localStorage.setItem("maidenmap.tile", '{"id":"does-not-exist"}');
    const { result } = renderHook(() => useMapTile("dark"));
    expect(result.current.provider.id).toBe("carto-dark");
  });
});
