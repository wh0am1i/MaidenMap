import { describe, expect, it } from "vitest";
import { qk } from "./query-keys";

describe("query keys", () => {
  it("grid single key is stable per code", () => {
    expect(qk.grid("JO65")).toEqual(["grid", "JO65"]);
  });

  it("grid-batch key sorts codes so [A,B] and [B,A] collide", () => {
    expect(qk.gridBatch(["B", "A"])).toEqual(["grid-batch", "A,B"]);
    expect(qk.gridBatch(["A", "B"])).toEqual(["grid-batch", "A,B"]);
  });

  it("health is a constant key", () => {
    expect(qk.health()).toEqual(["health"]);
  });
});
