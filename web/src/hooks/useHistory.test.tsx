import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { useHistory } from "./useHistory";

describe("useHistory", () => {
  beforeEach(() => localStorage.clear());

  it("starts empty", () => {
    const { result } = renderHook(() => useHistory());
    expect(result.current.entries).toEqual([]);
  });

  it("push prepends and dedupes to most recent", () => {
    const { result } = renderHook(() => useHistory());
    act(() => result.current.push({ grid: "JO65", label: "Malmö", countryCode: "SE", at: 1 }));
    act(() => result.current.push({ grid: "PM95", label: "Tsuru", countryCode: "JP", at: 2 }));
    act(() => result.current.push({ grid: "JO65", label: "Malmö", countryCode: "SE", at: 3 }));
    expect(result.current.entries.map((e) => e.grid)).toEqual(["JO65", "PM95"]);
    expect(result.current.entries[0].at).toBe(3);
  });

  it("caps at 20 entries", () => {
    const { result } = renderHook(() => useHistory());
    for (let i = 0; i < 25; i++) {
      act(() => result.current.push({ grid: `G${i}`, label: "", countryCode: "", at: i }));
    }
    expect(result.current.entries).toHaveLength(20);
    expect(result.current.entries[0].grid).toBe("G24");
    expect(result.current.entries[19].grid).toBe("G5");
  });

  it("clear empties list and storage", () => {
    const { result } = renderHook(() => useHistory());
    act(() => result.current.push({ grid: "JO65", label: "x", countryCode: "SE", at: 1 }));
    act(() => result.current.clear());
    expect(result.current.entries).toEqual([]);
    expect(localStorage.getItem("maidenmap.history")).toBeNull();
  });

  it("persists across hook instances", () => {
    const first = renderHook(() => useHistory());
    act(() => first.result.current.push({ grid: "JO65", label: "Malmö", countryCode: "SE", at: 1 }));
    first.unmount();
    const second = renderHook(() => useHistory());
    expect(second.result.current.entries.map((e) => e.grid)).toEqual(["JO65"]);
  });
});
