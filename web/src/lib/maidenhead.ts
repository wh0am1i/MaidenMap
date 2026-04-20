export interface GridBounds {
  latMin: number;
  latMax: number;
  lonMin: number;
  lonMax: number;
}

export interface ParsedGrid {
  kind: "ok";
  normalized: string;
  length: 4 | 6 | 8;
  lat: number;
  lon: number;
  bounds: GridBounds;
}

export type ParseError =
  | { kind: "error"; reason: "length"; message: string }
  | { kind: "error"; reason: "field"; message: string }
  | { kind: "error"; reason: "square"; message: string }
  | { kind: "error"; reason: "subsquare"; message: string }
  | { kind: "error"; reason: "extended"; message: string };

export type ParseResult = ParsedGrid | ParseError;

const FIELD_CHARS = "ABCDEFGHIJKLMNOPQR"; // A..R (18)
const DIGIT_CHARS = "0123456789";
const SUB_CHARS = "abcdefghijklmnopqrstuvwx"; // a..x (24)

export function parseMaidenhead(raw: string): ParseResult {
  const s = raw.trim();
  if (![4, 6, 8].includes(s.length)) {
    return { kind: "error", reason: "length", message: `invalid length ${s.length}: must be 4, 6, or 8` };
  }

  const c0 = s[0].toUpperCase();
  const c1 = s[1].toUpperCase();
  const c2 = s[2];
  const c3 = s[3];
  const c4 = s.length >= 6 ? s[4].toLowerCase() : null;
  const c5 = s.length >= 6 ? s[5].toLowerCase() : null;
  const c6 = s.length === 8 ? s[6] : null;
  const c7 = s.length === 8 ? s[7] : null;

  const f0 = FIELD_CHARS.indexOf(c0);
  const f1 = FIELD_CHARS.indexOf(c1);
  if (f0 < 0 || f1 < 0) return { kind: "error", reason: "field", message: `bad field character in ${s.slice(0, 2)}` };
  const d0 = DIGIT_CHARS.indexOf(c2);
  const d1 = DIGIT_CHARS.indexOf(c3);
  if (d0 < 0 || d1 < 0) return { kind: "error", reason: "square", message: `bad square digit in ${s.slice(2, 4)}` };

  let lonMin = -180 + f0 * 20 + d0 * 2;
  let latMin = -90 + f1 * 10 + d1 * 1;
  let lonStep = 2;
  let latStep = 1;
  let length: 4 | 6 | 8 = 4;
  let normalized = c0 + c1 + c2 + c3;

  if (c4 !== null && c5 !== null) {
    const s0 = SUB_CHARS.indexOf(c4);
    const s1 = SUB_CHARS.indexOf(c5);
    if (s0 < 0 || s1 < 0) {
      return { kind: "error", reason: "subsquare", message: `bad subsquare char in ${s.slice(4, 6)}` };
    }
    lonMin += s0 * (2 / 24);
    latMin += s1 * (1 / 24);
    lonStep = 2 / 24;
    latStep = 1 / 24;
    length = 6;
    normalized += c4 + c5;
  }

  if (c6 !== null && c7 !== null) {
    const e0 = DIGIT_CHARS.indexOf(c6);
    const e1 = DIGIT_CHARS.indexOf(c7);
    if (e0 < 0 || e1 < 0) {
      return { kind: "error", reason: "extended", message: `bad extended digit in ${s.slice(6, 8)}` };
    }
    lonMin += e0 * (lonStep / 10);
    latMin += e1 * (latStep / 10);
    lonStep = lonStep / 10;
    latStep = latStep / 10;
    length = 8;
    normalized += c6 + c7;
  }

  const lonMax = lonMin + lonStep;
  const latMax = latMin + latStep;
  return {
    kind: "ok",
    normalized,
    length,
    lat: round4((latMin + latMax) / 2),
    lon: round4((lonMin + lonMax) / 2),
    bounds: { latMin: round4(latMin), latMax: round4(latMax), lonMin: round4(lonMin), lonMax: round4(lonMax) },
  };
}

function round4(n: number): number {
  return Math.round(n * 10000) / 10000;
}
