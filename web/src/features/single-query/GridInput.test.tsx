import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { I18nextProvider } from "react-i18next";
import { describe, expect, it } from "vitest";
import i18n from "@/locales/i18n";
import { GridInput } from "./GridInput";

function wrap(node: React.ReactNode) {
  void i18n.changeLanguage("en");
  return render(<I18nextProvider i18n={i18n}>{node}</I18nextProvider>);
}

describe("GridInput", () => {
  it("calls onChange with normalized value on input", async () => {
    const user = userEvent.setup();
    const received: string[] = [];
    wrap(<GridInput value="" onChange={(v) => received.push(v)} />);
    const input = screen.getByRole("textbox");
    await user.type(input, "jo65ab");
    expect(received[received.length - 1]).toBe("JO65ab");
  });

  it("shows 'Valid' hint for 4-char correct input", () => {
    wrap(<GridInput value="JO65" onChange={() => {}} />);
    expect(screen.getByText(/Valid/i)).toBeInTheDocument();
  });

  it("shows 'need more' hint for 2-char input", () => {
    wrap(<GridInput value="JO" onChange={() => {}} />);
    expect(screen.getByText(/Need/i)).toBeInTheDocument();
  });

  it("shows error hint for bad field char", () => {
    wrap(<GridInput value="ZZ00" onChange={() => {}} />);
    expect(screen.getByText(/Invalid grid/i)).toBeInTheDocument();
  });

  it("calls onSubmit when pressing Enter on a valid value", async () => {
    const user = userEvent.setup();
    let submitted = "";
    wrap(<GridInput value="JO65" onChange={() => {}} onSubmit={(v) => (submitted = v)} />);
    const input = screen.getByRole("textbox");
    await user.click(input);
    await user.keyboard("{Enter}");
    expect(submitted).toBe("JO65");
  });
});
