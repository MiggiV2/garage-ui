import { describe, expect, it } from "vitest";
import { bucketName, formatBytes, parseSize, shortId } from "./format";

describe("formatBytes", () => {
  it("formats using binary units", () => {
    expect(formatBytes(0)).toBe("0 B");
    expect(formatBytes(1023)).toBe("1023 B");
    expect(formatBytes(1024)).toBe("1.0 KiB");
    expect(formatBytes(44721796)).toBe("42.7 MiB");
    expect(formatBytes(3000000000000)).toBe("2.7 TiB");
  });

  it("renders null as a dash", () => {
    expect(formatBytes(null)).toBe("–");
  });
});

describe("parseSize", () => {
  it("parses binary and decimal suffixes", () => {
    expect(parseSize("1024")).toBe(1024);
    expect(parseSize("10 GiB")).toBe(10 * 1024 ** 3);
    expect(parseSize("1.5GB")).toBe(1.5e9);
    expect(parseSize("2 tib")).toBe(2 * 1024 ** 4);
  });

  it("returns null for empty input", () => {
    expect(parseSize("")).toBeNull();
    expect(parseSize("   ")).toBeNull();
  });

  it("rejects garbage", () => {
    expect(() => parseSize("lots")).toThrow();
    expect(() => parseSize("-5 GB")).toThrow();
  });
});

describe("bucketName", () => {
  it("prefers the first global alias", () => {
    expect(bucketName({ id: "abcdef0123456789", globalAliases: ["velld", "other"], localAliases: [] })).toBe("velld");
  });

  it("falls back to a local alias, then to the short id", () => {
    expect(
      bucketName({ id: "abcdef0123456789", globalAliases: [], localAliases: [{ accessKeyId: "GK1", alias: "mine" }] }),
    ).toBe("mine");
    expect(bucketName({ id: "abcdef0123456789", globalAliases: [], localAliases: [] })).toBe("abcdef012345");
  });
});

describe("shortId", () => {
  it("truncates to 12 characters", () => {
    expect(shortId("07fc33fc32a1005618e58fa05e02fcfd")).toBe("07fc33fc32a1");
  });
});
