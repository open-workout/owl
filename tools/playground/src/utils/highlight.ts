export interface Token {
  text: string;
  className: string | null;
}

// Every literal keyword terminal in spec/grammar.ebnf — structural
// declarations, the sugar forms (dropset/drop/restpause/rest_pause),
// conditionals, and group constructs — get the same "keyword" color.
// See spec/grammar.ebnf for the source of truth; keep this list in
// sync with it rather than the other way around.
const OWL_KEYWORDS = [
  "state",
  "stats",
  "units",
  "plates",
  "block",
  "day",
  "exercise",
  "set",
  "progress",
  "dropset",
  "drop",
  "restpause",
  "rest_pause",
  "rest",
  "if",
  "then",
  "else",
  "and",
  "or",
  "superset",
  "circuit",
  "emom",
  "amrap",
  "for_time",
  "rounds",
  "as",
];

const OWL_TOKEN_RE = new RegExp(
  [
    `(?<comment>#.*$)`,
    `(?<string>"[^"]*"?)`,
    `(?<catalog>\\$[A-Za-z_][A-Za-z0-9_]*)`,
    `(?<number>\\d+(?:\\.\\d+)?(?:kg|lb|km|min|m|s|h|d)?\\+?)`,
    `(?<keyword>\\b(?:${OWL_KEYWORDS.join("|")})\\b)`,
    `(?<sentinel>\\b[A-Z][A-Z0-9_]*\\b)`,
    `(?<operator>[+\\-*/@=<>!]+)`,
  ].join("|"),
  "gm",
);

const JSON_TOKEN_RE = new RegExp(
  [
    `(?<key>"(?:[^"\\\\]|\\\\.)*"(?=\\s*:))`,
    `(?<string>"(?:[^"\\\\]|\\\\.)*")`,
    `(?<number>-?\\d+(?:\\.\\d+)?(?:[eE][+-]?\\d+)?)`,
    `(?<keyword>\\btrue\\b|\\bfalse\\b|\\bnull\\b)`,
    `(?<punct>[{}[\\]:,])`,
  ].join("|"),
  "gm",
);

function tokenize(line: string, re: RegExp): Token[] {
  const tokens: Token[] = [];
  let lastIndex = 0;
  re.lastIndex = 0;
  let match: RegExpExecArray | null;
  while ((match = re.exec(line))) {
    if (match.index > lastIndex) {
      tokens.push({ text: line.slice(lastIndex, match.index), className: null });
    }
    const groupName = Object.entries(match.groups ?? {}).find(([, value]) => value !== undefined)?.[0];
    tokens.push({ text: match[0], className: groupName ? `tok-${groupName}` : null });
    lastIndex = match.index + match[0].length;
    if (match[0].length === 0) re.lastIndex++; // guard against zero-width matches
  }
  if (lastIndex < line.length) {
    tokens.push({ text: line.slice(lastIndex), className: null });
  }
  return tokens;
}

export function highlightOwlLine(line: string): Token[] {
  return tokenize(line, OWL_TOKEN_RE);
}

export function highlightJsonLine(line: string): Token[] {
  return tokenize(line, JSON_TOKEN_RE);
}
