import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useTheme } from "./useTheme";

describe("useTheme", () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove("dark", "light");
  });
  afterEach(() => {
    document.documentElement.classList.remove("dark", "light");
  });

  it("follows system by default and sets dark class when system is dark", () => {
    const mql = { matches: true, media: "(prefers-color-scheme: dark)", addEventListener: vi.fn(), removeEventListener: vi.fn() };
    vi.spyOn(window, "matchMedia").mockReturnValue(mql as unknown as MediaQueryList);
    const { result } = renderHook(() => useTheme());
    expect(result.current.resolved).toBe("dark");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
  });

  it("user override to light wins over system", () => {
    const mql = { matches: true, media: "(prefers-color-scheme: dark)", addEventListener: vi.fn(), removeEventListener: vi.fn() };
    vi.spyOn(window, "matchMedia").mockReturnValue(mql as unknown as MediaQueryList);
    const { result } = renderHook(() => useTheme());
    act(() => result.current.setTheme("light"));
    expect(result.current.resolved).toBe("light");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    expect(localStorage.getItem("maidenmap.theme")).toBe("light");
  });

  it("restores persisted user override on mount", () => {
    localStorage.setItem("maidenmap.theme", "dark");
    const mql = { matches: false, media: "(prefers-color-scheme: dark)", addEventListener: vi.fn(), removeEventListener: vi.fn() };
    vi.spyOn(window, "matchMedia").mockReturnValue(mql as unknown as MediaQueryList);
    const { result } = renderHook(() => useTheme());
    expect(result.current.resolved).toBe("dark");
    expect(result.current.theme).toBe("dark");
  });
});
