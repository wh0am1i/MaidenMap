import { useEffect, useState } from "react";
import { Outlet, useNavigate, useRouteError, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Toaster } from "sonner";
import { TopBar, type TabKey } from "@/components/TopBar";

export default function Root() {
  const [params, setParams] = useSearchParams();
  const navigate = useNavigate();
  const activeTab = (params.get("tab") === "batch" ? "batch" : "single") as TabKey;

  function setTab(t: TabKey) {
    if (t === "single") params.delete("tab");
    else params.set("tab", "batch");
    setParams(params, { replace: true });
    navigate({ pathname: "/", search: params.toString() });
  }

  return (
    <div className="min-h-screen flex flex-col">
      <TopBar activeTab={activeTab} onTabChange={setTab} />
      <OfflineBanner />
      <main className="flex-1 min-h-0">
        <Outlet />
      </main>
      <Toaster position="top-right" />
    </div>
  );
}

function OfflineBanner() {
  const { t } = useTranslation();
  const [online, setOnline] = useState(() => (typeof navigator === "undefined" ? true : navigator.onLine));
  useEffect(() => {
    const up = () => setOnline(true);
    const down = () => setOnline(false);
    window.addEventListener("online", up);
    window.addEventListener("offline", down);
    return () => {
      window.removeEventListener("online", up);
      window.removeEventListener("offline", down);
    };
  }, []);
  if (online) return null;
  return (
    <div className="w-full px-3 py-1 text-[11px] text-center bg-[rgb(var(--danger))]/10 border-b border-[rgb(var(--danger))]/30 text-[rgb(var(--danger))]">
      {t("error.offline_banner")}
    </div>
  );
}

export function RootErrorBoundary() {
  const err = useRouteError();
  return (
    <div className="p-8">
      <h1 className="text-lg font-semibold">Something went wrong</h1>
      <pre className="mt-2 text-xs opacity-60">{String(err)}</pre>
    </div>
  );
}
