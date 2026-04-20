import type { ReactElement } from "react";
import { render, type RenderOptions } from "@testing-library/react";

export function renderWithProviders(ui: ReactElement, options?: RenderOptions) {
  return render(ui, options);
}

// This is a minimal wrapper for now. Query/Router/i18n providers are added
// once those modules exist (Tasks 5, 15, 25).
