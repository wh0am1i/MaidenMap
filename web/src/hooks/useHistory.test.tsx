import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { useHistory } from "./useHistory";

const mkBi = (en = "", zh = "") => ({ en, zh });

describe("useHistory", () => {
  beforeEach(() => localStorage.clear());

  it("starts empty", () => {
    const { result } = renderHook(() => useHistory());
    expect(result.current.entries).toEqual([]);
  });

  it("push prepends and dedupes to most recent", () => {
    const { result } = renderHook(() => useHistory());
    act(() => result.current.push({ grid: "JO65", label: mkBi("Malmö", "马尔默"), admin1: mkBi("Skåne", "斯科讷省"), countryCode: "SE", at: 1 }));
    act(() => result.current.push({ grid: "PM95", label: mkBi("Tsuru", ""), admin1: mkBi("Yamanashi", "山梨县"), countryCode: "JP", at: 2 }));
    act(() => result.current.push({ grid: "JO65", label: mkBi("Malmö", "马尔默"), admin1: mkBi("Skåne", "斯科讷省"), countryCode: "SE", at: 3 }));
    expect(result.current.entries.map((e) => e.grid)).toEqual(["JO65", "PM95"]);
    expect(result.current.entries[0].at).toBe(3);
  });

  it("caps at 20 entries", () => {
    const { result } = renderHook(() => useHistory());
    for (let i = 0; i < 25; i++) {
      act(() => result.current.push({ grid: `G${i}`, label: mkBi(), admin1: mkBi(), countryCode: "", at: i }));
    }
    expect(result.current.entries).toHaveLength(20);
    expect(result.current.entries[0].grid).toBe("G24");
    expect(result.current.entries[19].grid).toBe("G5");
  });

  it("remove drops a single entry by grid", () => {
    const { result } = renderHook(() => useHistory());
    act(() => result.current.push({ grid: "A", label: mkBi("a"), admin1: mkBi(), countryCode: "", at: 1 }));
    act(() => result.current.push({ grid: "B", label: mkBi("b"), admin1: mkBi(), countryCode: "", at: 2 }));
    act(() => result.current.remove("A"));
    expect(result.current.entries.map((e) => e.grid)).toEqual(["B"]);
  });

  it("clear empties list and storage", () => {
    const { result } = renderHook(() => useHistory());
    act(() => result.current.push({ grid: "JO65", label: mkBi("x"), admin1: mkBi(), countryCode: "SE", at: 1 }));
    act(() => result.current.clear());
    expect(result.current.entries).toEqual([]);
    expect(localStorage.getItem("maidenmap.history")).toBeNull();
  });

  it("persists across hook instances", () => {
    const first = renderHook(() => useHistory());
    act(() => first.result.current.push({ grid: "JO65", label: mkBi("Malmö", "马尔默"), admin1: mkBi("Skåne", "斯科讷省"), countryCode: "SE", at: 1 }));
    first.unmount();
    const second = renderHook(() => useHistory());
    expect(second.result.current.entries.map((e) => e.grid)).toEqual(["JO65"]);
    expect(second.result.current.entries[0].label).toEqual({ en: "Malmö", zh: "马尔默" });
  });

  it("migrates old string-label storage entries", () => {
    localStorage.setItem(
      "maidenmap.history",
      JSON.stringify([{ grid: "X", label: "legacy", countryCode: "US", at: 1 }]),
    );
    const { result } = renderHook(() => useHistory());
    expect(result.current.entries[0].label).toEqual({ en: "legacy", zh: "" });
    // Pre-admin1 entries fill in an empty admin1 rather than undefined — consumers
    // (History.tsx) detect "empty admin1" and fall back to label for rendering.
    expect(result.current.entries[0].admin1).toEqual({ en: "", zh: "" });
  });
});
