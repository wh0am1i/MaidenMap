import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { parseMaidenhead } from "@/lib/maidenhead";

function normalize(v: string): string {
  const s = v.trim();
  if (s.length <= 2) return s.toUpperCase();
  if (s.length <= 4) return s.slice(0, 2).toUpperCase() + s.slice(2);
  if (s.length <= 6) return s.slice(0, 2).toUpperCase() + s.slice(2, 4) + s.slice(4).toLowerCase();
  return s.slice(0, 2).toUpperCase() + s.slice(2, 4) + s.slice(4, 6).toLowerCase() + s.slice(6);
}

export function GridInput({
  value,
  onChange,
  onSubmit,
  autoFocus,
}: {
  value: string;
  onChange: (v: string) => void;
  onSubmit?: (v: string) => void;
  autoFocus?: boolean;
}) {
  const { t } = useTranslation();
  const [local, setLocal] = useState(value);

  useEffect(() => {
    setLocal(value);
  }, [value]);

  const display = local;

  const state = useMemo(() => {
    if (!display) return { status: "empty" as const };
    const parsed = parseMaidenhead(display);
    if (parsed.kind === "ok") return { status: "ok" as const, length: parsed.length };
    if (display.length < 4) return { status: "partial" as const, need: 4 - display.length };
    return { status: "error" as const, message: parsed.message };
  }, [display]);

  const hint =
    state.status === "empty"
      ? t("hint.enter_grid")
      : state.status === "ok"
        ? t("hint.valid_n_chars", { n: state.length })
        : state.status === "partial"
          ? t("hint.need_more_chars", { n: state.need })
          : t("error.invalid_grid", { message: state.message });

  const ring =
    state.status === "ok"
      ? "border-[rgb(var(--ham))]"
      : state.status === "error"
        ? "border-[rgb(var(--danger))]"
        : "border-border";

  return (
    <div className="flex flex-col gap-1">
      <Input
        autoFocus={autoFocus}
        value={display}
        onChange={(e) => {
          const n = normalize(e.target.value);
          setLocal(n);
          onChange(n);
        }}
        onKeyDown={(e) => {
          if (e.key === "Enter" && state.status === "ok" && onSubmit) onSubmit(display);
        }}
        placeholder={t("hint.placeholder_example")}
        className={cn("font-mono text-lg tracking-wider", ring)}
      />
      <p
        className={cn(
          "text-xs",
          state.status === "error" ? "text-[rgb(var(--danger))]" : "text-[rgb(var(--dim))]",
        )}
      >
        {hint}
      </p>
    </div>
  );
}
