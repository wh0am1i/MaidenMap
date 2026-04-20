import { useCallback, useEffect, useState } from "react";
import type { BiName } from "@/types/api";

export interface HistoryEntry {
  grid: string;
  label: BiName;
  countryCode: string;
  at: number; // ms epoch
}

const STORAGE_KEY = "maidenmap.history";
const CAP = 20;

function toBiName(v: unknown): BiName {
  if (typeof v === "string") return { en: v, zh: "" };
  if (v && typeof v === "object") {
    const o = v as Partial<BiName>;
    return { en: typeof o.en === "string" ? o.en : "", zh: typeof o.zh === "string" ? o.zh : "" };
  }
  return { en: "", zh: "" };
}

function load(): HistoryEntry[] {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed
      .filter((e): e is Record<string, unknown> => typeof e === "object" && e !== null && typeof e.grid === "string")
      .map((e) => ({
        grid: e.grid as string,
        label: toBiName(e.label),
        countryCode: typeof e.countryCode === "string" ? e.countryCode : "",
        at: typeof e.at === "number" ? e.at : 0,
      }));
  } catch {
    return [];
  }
}

function save(entries: HistoryEntry[]): void {
  if (entries.length === 0) localStorage.removeItem(STORAGE_KEY);
  else localStorage.setItem(STORAGE_KEY, JSON.stringify(entries));
}

export function useHistory(): {
  entries: HistoryEntry[];
  push: (entry: HistoryEntry) => void;
  remove: (grid: string) => void;
  clear: () => void;
} {
  const [entries, setEntries] = useState<HistoryEntry[]>(() => load());

  useEffect(() => {
    save(entries);
  }, [entries]);

  const push = useCallback((entry: HistoryEntry) => {
    setEntries((prev) => {
      const filtered = prev.filter((e) => e.grid !== entry.grid);
      return [entry, ...filtered].slice(0, CAP);
    });
  }, []);

  const remove = useCallback((grid: string) => {
    setEntries((prev) => prev.filter((e) => e.grid !== grid));
  }, []);

  const clear = useCallback(() => {
    setEntries([]);
  }, []);

  return { entries, push, remove, clear };
}
