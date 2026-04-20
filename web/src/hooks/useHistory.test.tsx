import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { useHistory } from "./useHistory";

const mkLabel = (en = "", zh = "") => ({ en, zh });

describe("useHistory", () => {
  beforeEach(() => localStorage.clear());

  it("starts empty", () => {
    const { result } = renderHook(() => useHistory());
    expect(result.current.entries).toEqual([]);
  });

  it("push prepends and dedupes to most recent", () => {
    const { result } = renderHook(() => useHistory());
    act(() => result.current.push({ grid: "JO65", label: mkLabel("Malmö", "马尔默"), countryCode: "SE", at: 1 }));
    act(() => result.current.push({ grid: "PM95", label: mkLabel("Tsuru", ""), countryCode: "JP", at: 2 }));
    act(() => result.current.push({ grid: "JO65", label: mkLabel("Malmö", "马尔默"), countryCode: "SE", at: 3 }));
    expect(result.current.entries.map((e) => e.grid)).toEqual(["JO65", "PM95"]);
    expect(result.current.entries[0].at).toBe(3);
  });

  it("caps at 20 entries", () => {
    const { result } = renderHook(() => useHistory());
    for (let i = 0; i < 25; i++) {
      act(() => result.current.push({ grid: `G${i}`, label: mkLabel(), countryCode: "", at: i }));
    }
    expect(result.current.entries).toHaveLength(20);
    expect(result.current.entries[0].grid).toBe("G24");
    expect(result.current.entries[19].grid).toBe("G5");
  });

  it("remove drops a single entry by grid", () => {
    const { result } = renderHook(() => useHistory());
    act(() => result.current.push({ grid: "A", label: mkLabel("a"), countryCode: "", at: 1 }));
    act(() => result.current.push({ grid: "B", label: mkLabel("b"), countryCode: "", at: 2 }));
    act(() => result.current.remove("A"));
    expect(result.current.entries.map((e) => e.grid)).toEqual(["B"]);
  });

  it("clear empties list and storage", () => {
    const { result } = renderHook(() => useHistory());
    act(() => result.current.push({ grid: "JO65", label: mkLabel("x"), countryCode: "SE", at: 1 }));
    act(() => result.current.clear());
    expect(result.current.entries).toEqual([]);
    expect(localStorage.getItem("maidenmap.history")).toBeNull();
  });

  it("persists across hook instances", () => {
    const first = renderHook(() => useHistory());
    act(() => first.result.current.push({ grid: "JO65", label: mkLabel("Malmö", "马尔默"), countryCode: "SE", at: 1 }));
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
  });
});
