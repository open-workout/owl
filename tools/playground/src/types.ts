export type FileKind = "owl" | "json" | "md";

export interface WorkspaceFile {
  id: string;
  name: string;
  kind: FileKind;
  content: string;
}

/** A view onto a file. `.md` files can be open as both `source` and `preview`. */
export type TabView = "source" | "preview";

export interface Tab {
  fileId: string;
  view: TabView;
}

export interface Pane {
  id: string;
  tabs: Tab[];
  activeTabIndex: number;
}
