import { useCallback, useSyncExternalStore } from "react";
import { builtInProviders, type TileProvider } from "@/lib/tile-providers";

const STORAGE_KEY = "maidenmap.tile";

interface Stored {
  id?: string;
  customUrl?: string;
}

function loadStored(): Stored {
  if (typeof localStorage === "undefined") return {};
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return {};
  try {
    return JSON.parse(raw) as Stored;
  } catch {
    return {};
  }
}

function persist(s: Stored): void {
  if (!s.id && !s.customUrl) localStorage.removeItem(STORAGE_KEY);
  else localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
}

// Module-level store so multiple useMapTile consumers (TopBar dropdown +
// Home map) stay in sync without threading props/context.
let current: Stored = loadStored();
const listeners = new Set<() => void>();
function subscribe(cb: () => void) {
  listeners.add(cb);
  return () => {
    listeners.delete(cb);
  };
}
function setStored(next: Stored) {
  current = next;
  persist(next);
  listeners.forEach((cb) => cb());
}

function defaultProviderFor(themeResolved: "light" | "dark"): TileProvider {
  const id = themeResolved === "dark" ? "carto-dark" : "carto-light";
  return builtInProviders.find((p) => p.id === id)!;
}

export function useMapTile(themeResolved: "light" | "dark"): {
  provider: TileProvider;
  setProviderId: (id: string) => void;
  setCustomUrl: (url: string) => void;
  clear: () => void;
} {
  const stored = useSyncExternalStore(
    subscribe,
    () => current,
    () => current,
  );

  let provider: TileProvider;
  if (stored.customUrl) {
    provider = {
      id: "custom",
      name: "Custom",
      url: stored.customUrl,
      attribution: "Custom tile provider",
      variant: "any",
      maxZoom: 19,
    };
  } else if (stored.id) {
    const found = builtInProviders.find((p) => p.id === stored.id);
    provider = found ?? defaultProviderFor(themeResolved);
  } else {
    provider = defaultProviderFor(themeResolved);
  }

  const setProviderId = useCallback((id: string) => setStored({ id }), []);
  const setCustomUrl = useCallback((url: string) => setStored({ customUrl: url }), []);
  const clear = useCallback(() => setStored({}), []);

  return { provider, setProviderId, setCustomUrl, clear };
}

// Test helper: reinitialize the module store from localStorage
// (tests clear localStorage in beforeEach; this picks that up).
export function __resetMapTileStore() {
  current = loadStored();
  listeners.forEach((cb) => cb());
}
