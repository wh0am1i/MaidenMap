import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { I18nextProvider } from "react-i18next";
import { describe, expect, it, vi } from "vitest";
import i18n from "@/locales/i18n";
import { ResultCard } from "./ResultCard";
import type { GridResponse } from "@/types/api";

const data: GridResponse = {
  grid: "JO65ab",
  center: { lat: 55.0625, lon: 12.0417 },
  country: { code: "DK", name: { en: "Denmark", zh: "丹麦" } },
  admin1: { en: "Zealand", zh: "西兰" },
  admin2: { en: "Vordingborg Kommune", zh: "" },
  city: { en: "Vordingborg", zh: "" },
};

function wrap(node: React.ReactNode) {
  void i18n.changeLanguage("en");
  return render(<I18nextProvider i18n={i18n}>{node}</I18nextProvider>);
}

describe("ResultCard", () => {
  it("shows grid code and center coords", () => {
    wrap(<ResultCard data={data} />);
    expect(screen.getByText("JO65ab")).toBeInTheDocument();
    expect(screen.getByText(/55.0625/)).toBeInTheDocument();
    expect(screen.getByText(/12.0417/)).toBeInTheDocument();
  });

  it("copies a bilingual summary to clipboard on click", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText },
      configurable: true,
    });
    wrap(<ResultCard data={data} />);
    await user.click(screen.getByRole("button", { name: /Copy/i }));
    expect(writeText).toHaveBeenCalledOnce();
    const arg = writeText.mock.calls[0][0] as string;
    expect(arg).toMatch(/JO65ab/);
    // Both EN and ZH appear where available
    expect(arg).toMatch(/Denmark/);
    expect(arg).toMatch(/丹麦/);
    expect(arg).toMatch(/Zealand/);
    expect(arg).toMatch(/西兰/);
    expect(arg).toMatch(/Vordingborg/);
    // Smart N/E signs
    expect(arg).toMatch(/55\.0625°N/);
    expect(arg).toMatch(/12\.0417°E/);
  });

  it("formats southern / western coordinates with S / W", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText },
      configurable: true,
    });
    wrap(<ResultCard data={{ ...data, center: { lat: -34.5, lon: -58.5 } }} />);
    await user.click(screen.getByRole("button", { name: /Copy/i }));
    const arg = writeText.mock.calls[0][0] as string;
    expect(arg).toMatch(/34\.5°S/);
    expect(arg).toMatch(/58\.5°W/);
  });

  it("renders null country gracefully", () => {
    wrap(<ResultCard data={{ ...data, country: null }} />);
    expect(screen.getByText(/JO65ab/)).toBeInTheDocument();
  });
});
