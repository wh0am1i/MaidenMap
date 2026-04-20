import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { I18nextProvider } from "react-i18next";
import { beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/locales/i18n";
import { TopBar } from "./TopBar";

beforeEach(() => {
  const mql = {
    matches: false,
    media: "(prefers-color-scheme: dark)",
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  };
  vi.spyOn(window, "matchMedia").mockReturnValue(mql as unknown as MediaQueryList);
});

function wrap(node: React.ReactNode) {
  const qc = new QueryClient();
  return (
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <I18nextProvider i18n={i18n}>{node}</I18nextProvider>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe("TopBar", () => {
  it("renders brand", () => {
    render(wrap(<TopBar activeTab="single" onTabChange={() => {}} />));
    expect(screen.getByText(/MaidenMap/i)).toBeInTheDocument();
  });

  it("language switcher toggles i18n language", async () => {
    const user = userEvent.setup();
    void i18n.changeLanguage("en");
    render(wrap(<TopBar activeTab="single" onTabChange={() => {}} />));
    const langBtn = screen.getByRole("button", { name: /中/i });
    await user.click(langBtn);
    expect(i18n.language).toBe("zh-CN");
  });

  it("calls onTabChange when clicking a tab button", async () => {
    const user = userEvent.setup();
    const calls: string[] = [];
    render(wrap(<TopBar activeTab="single" onTabChange={(t) => calls.push(t)} />));
    const batchBtn = screen.getAllByRole("button").find((b) => /batch|批量/i.test(b.textContent ?? ""))!;
    await user.click(batchBtn);
    expect(calls).toEqual(["batch"]);
  });
});
