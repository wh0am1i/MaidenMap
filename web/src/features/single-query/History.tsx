import { useTranslation } from "react-i18next";
import { Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { HistoryEntry } from "@/hooks/useHistory";

// Re-export under a friendlier name for consumers that want to talk about
// rendered rows rather than stored entries. Shape is identical.
export type HistoryItem = HistoryEntry;

export function History({
  items,
  onPick,
  onClear,
}: {
  items: HistoryItem[];
  onPick: (grid: string) => void;
  onClear: () => void;
}) {
  const { t, i18n } = useTranslation();
  const rtf = new Intl.RelativeTimeFormat(i18n.language, { style: "short" });

  return (
    <div className="bg-[rgb(var(--panel-2))] border border-border rounded-lg overflow-hidden">
      <div className="flex items-center justify-between px-3 py-2 border-b border-border text-xs uppercase tracking-wider text-[rgb(var(--dim))]">
        <span>{t("history.title")}</span>
        {items.length > 0 && (
          <Button size="sm" variant="ghost" aria-label={t("history.clear_all")} onClick={onClear}>
            <Trash2 size={12} />
          </Button>
        )}
      </div>
      {items.length === 0 ? (
        <p className="px-3 py-4 text-xs text-[rgb(var(--dim))] text-center">{t("history.empty")}</p>
      ) : (
        <ul>
          {items.map((it) => (
            <li key={it.grid}>
              <button
                type="button"
                onClick={() => onPick(it.grid)}
                className="w-full flex items-center gap-2 px-3 py-[6px] text-left text-xs hover:bg-white/5"
              >
                <span className="font-mono text-[rgb(var(--ham))] w-[70px]">{it.grid}</span>
                <span className="flex-1 truncate text-[rgb(var(--dim))]">{it.label}</span>
                <span className="text-[10px] text-[rgb(var(--dimmer))]">{rtf.format(-secondsAgo(it.at), "second") || ""}</span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function secondsAgo(ts: number): number {
  return Math.round((Date.now() - ts) / 1000);
}
