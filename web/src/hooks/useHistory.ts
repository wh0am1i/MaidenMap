import { useCallback, useEffect, useState } from "react";

export interface HistoryEntry {
  grid: string;
  label: string;
  countryCode: string;
  at: number; // ms epoch
}

const STORAGE_KEY = "maidenmap.history";
const CAP = 20;

function load(): HistoryEntry[] {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(
      (e): e is HistoryEntry =>
        typeof e === "object" &&
        e !== null &&
        typeof (e as HistoryEntry).grid === "string",
    );
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

  const clear = useCallback(() => {
    setEntries([]);
  }, []);

  return { entries, push, clear };
}
