import { useTranslation } from "react-i18next";
import { Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { LocalizedName } from "@/components/LocalizedName";
import type { HistoryEntry } from "@/hooks/useHistory";

// Re-export under a friendlier name for consumers that want to talk about
// rendered rows rather than stored entries. Shape is identical.
export type HistoryItem = HistoryEntry;

export function History({
  items,
  onPick,
  onRemove,
  onClear,
}: {
  items: HistoryItem[];
  onPick: (grid: string) => void;
  onRemove?: (grid: string) => void;
  onClear: () => void;
}) {
  const { t, i18n } = useTranslation();

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
            <li key={it.grid} className="group flex items-center gap-1 px-3 py-[6px] hover:bg-white/5">
              <button
                type="button"
                onClick={() => onPick(it.grid)}
                className="flex-1 flex items-center gap-2 text-left text-xs min-w-0"
              >
                <span className="font-mono text-[rgb(var(--ham))] w-[70px] shrink-0">{it.grid}</span>
                <LocalizedName value={secondary(it)} className="flex-1 truncate text-[rgb(var(--dim))]" />
                <span className="text-[10px] text-[rgb(var(--dimmer))] shrink-0">{formatRelative(it.at, i18n.language)}</span>
              </button>
              {onRemove && (
                <button
                  type="button"
                  aria-label={t("action.remove")}
                  onClick={(e) => {
                    e.stopPropagation();
                    onRemove(it.grid);
                  }}
                  className="opacity-0 group-hover:opacity-60 hover:!opacity-100 text-[rgb(var(--dim))] p-[2px]"
                >
                  <X size={12} />
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

// Prefer admin1 (省/州). Legacy entries persisted before admin1 was tracked
// fall back to the old label so the list doesn't go blank for them.
function secondary(it: HistoryItem) {
  if (it.admin1 && (it.admin1.en || it.admin1.zh)) return it.admin1;
  return it.label;
}

function formatRelative(ts: number, lang: string): string {
  const diffSec = (Date.now() - ts) / 1000;
  const rtf = new Intl.RelativeTimeFormat(lang, { style: "short" });
  if (diffSec < 60) return rtf.format(-Math.round(diffSec), "second");
  if (diffSec < 3600) return rtf.format(-Math.round(diffSec / 60), "minute");
  if (diffSec < 86400) return rtf.format(-Math.round(diffSec / 3600), "hour");
  return rtf.format(-Math.round(diffSec / 86400), "day");
}
