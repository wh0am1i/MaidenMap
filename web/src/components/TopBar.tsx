import { useTranslation } from "react-i18next";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Moon, Sun, Monitor, Layers } from "lucide-react";
import { cn } from "@/lib/utils";
import { qk } from "@/lib/query-keys";
import { getHealth } from "@/lib/api";
import { useTheme } from "@/hooks/useTheme";
import { useMapTile } from "@/hooks/useMapTile";
import { builtInProviders } from "@/lib/tile-providers";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export type TabKey = "single" | "batch";

export function TopBar({
  activeTab,
  onTabChange,
}: {
  activeTab: TabKey;
  onTabChange: (t: TabKey) => void;
}) {
  const { t, i18n } = useTranslation();
  const { theme, resolved, setTheme } = useTheme();
  const tile = useMapTile(resolved);
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
        <Link
          to="/about"
          className="px-2 py-[2px] border border-border rounded hover:text-[rgb(var(--text))]"
        >
          {t("nav.about")}
        </Link>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              type="button"
              className="p-[3px] border border-border rounded inline-flex items-center"
              aria-label={t("action.map_layer")}
            >
              <Layers size={14} />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>{t("action.map_layer")}</DropdownMenuLabel>
            <DropdownMenuSeparator />
            {builtInProviders.map((p) => (
              <DropdownMenuItem
                key={p.id}
                onSelect={() => tile.setProviderId(p.id)}
                className={cn(tile.provider.id === p.id && "font-semibold text-[rgb(var(--ham))]")}
              >
                {p.name}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
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
