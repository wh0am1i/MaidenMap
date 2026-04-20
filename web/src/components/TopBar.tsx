import { useTranslation } from "react-i18next";
import { useQuery } from "@tanstack/react-query";
import { Moon, Sun, Monitor } from "lucide-react";
import { cn } from "@/lib/utils";
import { qk } from "@/lib/query-keys";
import { getHealth } from "@/lib/api";
import { useTheme } from "@/hooks/useTheme";

export type TabKey = "single" | "batch";

export function TopBar({
  activeTab,
  onTabChange,
}: {
  activeTab: TabKey;
  onTabChange: (t: TabKey) => void;
}) {
  const { t, i18n } = useTranslation();
  const { theme, setTheme } = useTheme();
  const health = useQuery({ queryKey: qk.health(), queryFn: getHealth, refetchInterval: 60_000 });

  return (
    <header className="flex items-center justify-between px-4 py-2 border-b border-border bg-[rgb(var(--panel))]">
      <div className="flex items-center gap-2 font-semibold">
        <span className="w-6 h-6 rounded bg-[rgb(var(--ham))] text-white inline-flex items-center justify-center text-[11px]">FM</span>
        <span>{t("brand.name")}</span>
        <span className="text-xs text-[rgb(var(--dim))] font-mono ml-1">· {t("brand.tagline")}</span>
      </div>

      <div className="inline-flex gap-1 p-[3px] bg-[rgb(var(--panel-2))] rounded-md border border-border">
        {(["single", "batch"] as TabKey[]).map((tab) => (
          <button
            key={tab}
            type="button"
            onClick={() => onTabChange(tab)}
            className={cn(
              "px-3 py-1 text-xs rounded",
              activeTab === tab
                ? "bg-[rgb(var(--bg))] text-[rgb(var(--text))] ring-1 ring-border"
                : "text-[rgb(var(--dim))]",
            )}
          >
            {t(`nav.${tab}`)}
          </button>
        ))}
      </div>

      <div className="flex items-center gap-3 text-xs text-[rgb(var(--dim))]">
        {health.data && (
          <span className="font-mono px-2 py-[2px] border border-border rounded">
            {(health.data.cities_count / 1000).toFixed(1)}k cities
          </span>
        )}
        <button
          type="button"
          className="px-2 py-[2px] border border-border rounded"
          onClick={() => i18n.changeLanguage(i18n.language.startsWith("zh") ? "en" : "zh-CN")}
        >
          {i18n.language.startsWith("zh") ? "EN" : "中"}
        </button>
        <button
          type="button"
          className="p-[3px] border border-border rounded"
          aria-label="theme"
          onClick={() => setTheme(theme === "light" ? "dark" : theme === "dark" ? "system" : "light")}
        >
          {theme === "light" ? <Sun size={14} /> : theme === "dark" ? <Moon size={14} /> : <Monitor size={14} />}
        </button>
      </div>
    </header>
  );
}
