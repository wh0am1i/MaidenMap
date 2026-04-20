import { describe, expect, it } from "vitest";
import { toCSV } from "./csv";

describe("toCSV", () => {
  it("escapes quotes and commas", () => {
    const csv = toCSV(
      ["name", "note"],
      [
        ["Alice", 'needs, "quoting"'],
        ["Bob", "plain"],
      ],
    );
    expect(csv).toBe(`name,note\nAlice,"needs, ""quoting"""\nBob,plain`);
  });

  it("emits header only when rows is empty", () => {
    expect(toCSV(["a", "b"], [])).toBe("a,b");
  });

  it("handles newlines inside cells", () => {
    const csv = toCSV(["x"], [["line1\nline2"]]);
    expect(csv).toBe(`x\n"line1\nline2"`);
  });
});
