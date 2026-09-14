import type { FileKind, Pane, Tab, TabView, WorkspaceFile } from "../types";

const STORAGE_KEY = "owl-playground:workspace";
const STORAGE_VERSION = 1;

export interface PersistedWorkspace {
  files: WorkspaceFile[];
  panes: Pane[];
  focusedPaneId: string;
}

const FILE_KINDS: FileKind[] = ["owl", "json", "md"];
const TAB_VIEWS: TabView[] = ["source", "preview"];

function isFile(value: unknown): value is WorkspaceFile {
  if (typeof value !== "object" || value === null) return false;
  const f = value as Record<string, unknown>;
  return (
    typeof f.id === "string" &&
    typeof f.name === "string" &&
    typeof f.content === "string" &&
    FILE_KINDS.includes(f.kind as FileKind)
  );
}

function sanitizeTabs(tabs: unknown, fileIds: Set<string>): Tab[] {
  if (!Array.isArray(tabs)) return [];
  const result: Tab[] = [];
  for (const t of tabs) {
    if (typeof t !== "object" || t === null) continue;
    const tab = t as Record<string, unknown>;
    if (typeof tab.fileId !== "string" || !fileIds.has(tab.fileId)) continue;
    if (!TAB_VIEWS.includes(tab.view as TabView)) continue;
    result.push({ fileId: tab.fileId, view: tab.view as TabView });
  }
  return result;
}

function sanitizePanes(panes: unknown, fileIds: Set<string>): Pane[] {
  if (!Array.isArray(panes) || panes.length === 0) return [];
  const result: Pane[] = [];
  for (const p of panes) {
    if (typeof p !== "object" || p === null) continue;
    const pane = p as Record<string, unknown>;
    if (typeof pane.id !== "string") continue;
    const tabs = sanitizeTabs(pane.tabs, fileIds);
    const activeTabIndex =
      typeof pane.activeTabIndex === "number" && pane.activeTabIndex >= 0 && pane.activeTabIndex < tabs.length
        ? pane.activeTabIndex
        : tabs.length > 0
          ? 0
          : -1;
    const size = typeof pane.size === "number" && pane.size > 0 ? pane.size : 100 / panes.length;
    result.push({ id: pane.id, tabs, activeTabIndex, size });
  }
  return result;
}

/** Reads and validates the persisted workspace; returns null if there's
 * nothing stored, storage is unavailable, or the stored data doesn't hold
 * up (wrong version, corrupted JSON, references that don't resolve). */
export function loadWorkspace(): PersistedWorkspace | null {
  let raw: string | null;
  try {
    raw = localStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
  if (!raw) return null;

  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }
  if (typeof parsed !== "object" || parsed === null) return null;
  const data = parsed as Record<string, unknown>;
  if (data.version !== STORAGE_VERSION) return null;

  const files = Array.isArray(data.files) ? data.files.filter(isFile) : [];
  const fileIds = new Set(files.map((f) => f.id));
  const panes = sanitizePanes(data.panes, fileIds);
  if (panes.length === 0) return null;

  const focusedPaneId =
    typeof data.focusedPaneId === "string" && panes.some((p) => p.id === data.focusedPaneId)
      ? data.focusedPaneId
      : panes[0].id;

  return { files, panes, focusedPaneId };
}

export function saveWorkspace(workspace: PersistedWorkspace): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ version: STORAGE_VERSION, ...workspace }));
  } catch {
    // Storage full or unavailable (private browsing, quota) — the
    // session keeps working in memory, it just won't survive a reload.
  }
}
