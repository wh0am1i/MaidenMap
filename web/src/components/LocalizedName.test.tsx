import { render, screen } from "@testing-library/react";
import { I18nextProvider } from "react-i18next";
import { describe, expect, it } from "vitest";
import i18n from "@/locales/i18n";
import { LocalizedName } from "./LocalizedName";

function renderWithLang(lang: string, node: React.ReactNode) {
  void i18n.changeLanguage(lang);
  return render(<I18nextProvider i18n={i18n}>{node}</I18nextProvider>);
}

describe("LocalizedName", () => {
  it("picks zh when language is zh-CN", () => {
    renderWithLang("zh-CN", <LocalizedName value={{ en: "Denmark", zh: "丹麦" }} />);
    expect(screen.getByText("丹麦")).toBeInTheDocument();
  });

  it("picks en when language is en", () => {
    renderWithLang("en", <LocalizedName value={{ en: "Denmark", zh: "丹麦" }} />);
    expect(screen.getByText("Denmark")).toBeInTheDocument();
  });

  it("falls back when zh is empty under zh-CN", () => {
    renderWithLang("zh-CN", <LocalizedName value={{ en: "OnlyEn", zh: "" }} />);
    expect(screen.getByText("OnlyEn")).toBeInTheDocument();
  });

  it("falls back when en is empty under en", () => {
    renderWithLang("en", <LocalizedName value={{ en: "", zh: "仅中文" }} />);
    expect(screen.getByText("仅中文")).toBeInTheDocument();
  });

  it("shows placeholder when both empty", () => {
    renderWithLang("en", <LocalizedName value={{ en: "", zh: "" }} />);
    expect(screen.getByText("—")).toBeInTheDocument();
  });
});
