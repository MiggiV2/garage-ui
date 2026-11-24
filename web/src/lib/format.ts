import type { BucketSummary } from "./types";

const BINARY_UNITS = ["B", "KiB", "MiB", "GiB", "TiB", "PiB"];

const SIZE_FACTORS: Record<string, number> = {
  "": 1,
  b: 1,
  k: 1e3,
  kb: 1e3,
  m: 1e6,
  mb: 1e6,
  g: 1e9,
  gb: 1e9,
  t: 1e12,
  tb: 1e12,
  kib: 1024,
  mib: 1024 ** 2,
  gib: 1024 ** 3,
  tib: 1024 ** 4,
};

export function formatBytes(bytes: number | null): string {
  if (bytes === null) return "–";
  if (bytes < 1024) return `${bytes} B`;
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < BINARY_UNITS.length - 1) {
    value /= 1024;
    unit++;
  }
  return `${value.toFixed(1)} ${BINARY_UNITS[unit]}`;
}

export function parseSize(input: string): number | null {
  const trimmed = input.trim();
  if (trimmed === "") return null;
  const match = /^(\d+(?:[.,]\d+)?)\s*([a-z]*)$/i.exec(trimmed);
  const factor = match ? SIZE_FACTORS[match[2].toLowerCase()] : undefined;
  if (!match || factor === undefined) {
    throw new Error(`Invalid size "${input}" — use e.g. 500 GiB or 1.5 TB`);
  }
  return Math.round(Number(match[1].replace(",", ".")) * factor);
}

export function bucketName(bucket: Pick<BucketSummary, "id" | "globalAliases" | "localAliases">): string {
  return bucket.globalAliases[0] ?? bucket.localAliases[0]?.alias ?? shortId(bucket.id);
}

export function shortId(id: string): string {
  return id.slice(0, 12);
}
