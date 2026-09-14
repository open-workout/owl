import type { FileKind } from "../types";

const EXTENSION_TO_KIND: Record<string, FileKind> = {
  owl: "owl",
  json: "json",
  md: "md",
};

export const FILE_KIND_ICON: Record<FileKind, string> = {
  owl: "📄",
  json: "🧾",
  md: "📝",
};

export function kindFromFileName(name: string): FileKind | null {
  const extension = name.trim().split(".").pop()?.toLowerCase();
  if (!extension) return null;
  return EXTENSION_TO_KIND[extension] ?? null;
}

export function defaultContentForKind(kind: FileKind): string {
  switch (kind) {
    case "owl":
      return "block main = {\n\n}\n";
    case "json":
      return "{}\n";
    case "md":
      return "# Untitled\n";
  }
}
