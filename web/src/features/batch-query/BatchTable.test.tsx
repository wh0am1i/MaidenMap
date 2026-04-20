import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { I18nextProvider } from "react-i18next";
import { describe, expect, it } from "vitest";
import i18n from "@/locales/i18n";
import { BatchTable, type BatchRow } from "./BatchTable";

function wrap(node: React.ReactNode) {
  void i18n.changeLanguage("en");
  return render(<I18nextProvider i18n={i18n}>{node}</I18nextProvider>);
}

const rows: BatchRow[] = [
  { grid: "JO65", kind: "ok", country: { code: "SE", nameEn: "Sweden" }, label: "Malmö", seq: 1 },
  { grid: "BAD", kind: "error", message: "invalid" },
];

describe("BatchTable", () => {
  it("renders ok and error rows", () => {
    wrap(<BatchTable rows={rows} activeGrid={null} onRowClick={() => {}} />);
    expect(screen.getByText("JO65")).toBeInTheDocument();
    expect(screen.getByText(/invalid/)).toBeInTheDocument();
  });

  it("calls onRowClick with grid code", async () => {
    const user = userEvent.setup();
    const picks: string[] = [];
    wrap(<BatchTable rows={rows} activeGrid={null} onRowClick={(g) => picks.push(g)} />);
    await user.click(screen.getByText("JO65"));
    expect(picks).toEqual(["JO65"]);
  });

  it("highlights active grid", () => {
    wrap(<BatchTable rows={rows} activeGrid="JO65" onRowClick={() => {}} />);
    const row = screen.getByText("JO65").closest("tr")!;
    expect(row).toHaveAttribute("data-active", "true");
  });
});
