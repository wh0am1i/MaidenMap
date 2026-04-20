import { describe, expect, it, beforeEach } from "vitest";
import i18n from "./i18n";

describe("i18n", () => {
  beforeEach(async () => {
    await i18n.changeLanguage("en");
  });

  it("defaults to en with the Enter hint key", () => {
    expect(i18n.t("hint.enter_grid")).toBe("Enter a Maidenhead grid");
  });

  it("switches to zh-CN", async () => {
    await i18n.changeLanguage("zh-CN");
    expect(i18n.t("hint.enter_grid")).toBe("请输入梅登海德网格");
  });

  it("falls back to key on missing translation", () => {
    expect(i18n.t("this.key.does.not.exist")).toBe("this.key.does.not.exist");
  });
});
