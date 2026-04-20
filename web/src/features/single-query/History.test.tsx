import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { I18nextProvider } from "react-i18next";
import { describe, expect, it } from "vitest";
import i18n from "@/locales/i18n";
import { History, type HistoryItem } from "./History";

function wrap(node: React.ReactNode) {
  void i18n.changeLanguage("en");
  return render(<I18nextProvider i18n={i18n}>{node}</I18nextProvider>);
}

const items: HistoryItem[] = [
  { grid: "JO65", label: "Malmö", countryCode: "SE", at: Date.now() },
  { grid: "PM95", label: "Tsuru", countryCode: "JP", at: Date.now() - 60_000 },
];

describe("History", () => {
  it("renders empty state", () => {
    wrap(<History items={[]} onPick={() => {}} onClear={() => {}} />);
    expect(screen.getByText(/No recent/i)).toBeInTheDocument();
  });

  it("renders items and calls onPick on click", async () => {
    const user = userEvent.setup();
    const picked: string[] = [];
    wrap(<History items={items} onPick={(g) => picked.push(g)} onClear={() => {}} />);
    await user.click(screen.getByText("JO65"));
    expect(picked).toEqual(["JO65"]);
  });

  it("calls onClear when clicking clear", async () => {
    const user = userEvent.setup();
    let cleared = false;
    wrap(<History items={items} onPick={() => {}} onClear={() => (cleared = true)} />);
    await user.click(screen.getByRole("button", { name: /Clear history/i }));
    expect(cleared).toBe(true);
  });
});
