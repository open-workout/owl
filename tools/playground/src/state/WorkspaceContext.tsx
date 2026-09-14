import {
  createContext,
  useCallback,
  useContext,
  useReducer,
  type ReactNode,
} from "react";
import type { FileKind, Pane, Tab, WorkspaceFile } from "../types";
import { defaultContentForKind } from "../utils/fileTypes";

const MAX_PANES = 3;
export const MIN_PANE_SIZE = 15;

function makeId(): string {
  return crypto.randomUUID();
}

function makePane(size: number): Pane {
  return { id: makeId(), tabs: [], activeTabIndex: -1, size };
}

/** Reset every pane to an equal share of the row (weights summing to 100). */
function withEqualSizes(panes: Pane[]): Pane[] {
  const size = 100 / panes.length;
  return panes.map((pane) => ({ ...pane, size }));
}

interface WorkspaceState {
  files: WorkspaceFile[];
  panes: Pane[];
  focusedPaneId: string;
}

function initialState(): WorkspaceState {
  const pane = makePane(100);
  return { files: [], panes: [pane], focusedPaneId: pane.id };
}

type Action =
  | { type: "CREATE_FILE"; name: string; kind: FileKind }
  | { type: "RENAME_FILE"; id: string; name: string }
  | { type: "DELETE_FILE"; id: string }
  | { type: "UPDATE_FILE_CONTENT"; id: string; content: string }
  | { type: "OPEN_FILE"; fileId: string; paneId: string }
  | { type: "CLOSE_TAB"; paneId: string; tabIndex: number }
  | { type: "ADD_PANE" }
  | { type: "REMOVE_PANE"; paneId: string }
  | { type: "SET_FOCUSED_PANE"; paneId: string }
  | { type: "SET_ACTIVE_TAB"; paneId: string; tabIndex: number }
  | { type: "RESIZE_PANES"; leftPaneId: string; rightPaneId: string; leftSize: number; rightSize: number };

function tabsForFile(kind: FileKind, fileId: string): Tab[] {
  return kind === "md"
    ? [
        { fileId, view: "source" },
        { fileId, view: "preview" },
      ]
    : [{ fileId, view: "source" }];
}

function openFileInPane(pane: Pane, fileId: string, kind: FileKind): Pane {
  const existingIndex = pane.tabs.findIndex((tab) => tab.fileId === fileId);
  if (existingIndex !== -1) {
    return { ...pane, activeTabIndex: existingIndex };
  }
  const newTabs = [...pane.tabs, ...tabsForFile(kind, fileId)];
  return { ...pane, tabs: newTabs, activeTabIndex: pane.tabs.length };
}

function reducer(state: WorkspaceState, action: Action): WorkspaceState {
  switch (action.type) {
    case "CREATE_FILE": {
      const file: WorkspaceFile = {
        id: makeId(),
        name: action.name,
        kind: action.kind,
        content: defaultContentForKind(action.kind),
      };
      const panes = state.panes.map((pane) =>
        pane.id === state.focusedPaneId
          ? openFileInPane(pane, file.id, file.kind)
          : pane,
      );
      return { ...state, files: [...state.files, file], panes };
    }

    case "RENAME_FILE": {
      const name = action.name.trim();
      if (!name) return state;
      return {
        ...state,
        files: state.files.map((file) =>
          file.id === action.id ? { ...file, name } : file,
        ),
      };
    }

    case "DELETE_FILE": {
      const files = state.files.filter((file) => file.id !== action.id);
      const panes = state.panes.map((pane) => {
        const tabs = pane.tabs.filter((tab) => tab.fileId !== action.id);
        if (tabs.length === pane.tabs.length) return pane;
        const activeTabIndex = tabs.length === 0 ? -1 : Math.min(pane.activeTabIndex, tabs.length - 1);
        return { ...pane, tabs, activeTabIndex };
      });
      return { ...state, files, panes };
    }

    case "UPDATE_FILE_CONTENT": {
      return {
        ...state,
        files: state.files.map((file) =>
          file.id === action.id ? { ...file, content: action.content } : file,
        ),
      };
    }

    case "OPEN_FILE": {
      const file = state.files.find((f) => f.id === action.fileId);
      if (!file) return state;
      const panes = state.panes.map((pane) =>
        pane.id === action.paneId
          ? openFileInPane(pane, file.id, file.kind)
          : pane,
      );
      return { ...state, panes, focusedPaneId: action.paneId };
    }

    case "CLOSE_TAB": {
      const panes = state.panes.map((pane) => {
        if (pane.id !== action.paneId) return pane;
        const tabs = pane.tabs.filter((_, index) => index !== action.tabIndex);
        let activeTabIndex = pane.activeTabIndex;
        if (tabs.length === 0) {
          activeTabIndex = -1;
        } else if (action.tabIndex <= pane.activeTabIndex) {
          activeTabIndex = Math.max(0, pane.activeTabIndex - 1);
        }
        return { ...pane, tabs, activeTabIndex };
      });
      return { ...state, panes };
    }

    case "ADD_PANE": {
      if (state.panes.length >= MAX_PANES) return state;
      const newPane = makePane(0);
      const panes = withEqualSizes([...state.panes, newPane]);
      return { ...state, panes, focusedPaneId: newPane.id };
    }

    case "REMOVE_PANE": {
      if (state.panes.length <= 1) return state;
      const panes = withEqualSizes(state.panes.filter((pane) => pane.id !== action.paneId));
      const focusedPaneId =
        state.focusedPaneId === action.paneId ? panes[0].id : state.focusedPaneId;
      return { ...state, panes, focusedPaneId };
    }

    case "SET_FOCUSED_PANE":
      return { ...state, focusedPaneId: action.paneId };

    case "SET_ACTIVE_TAB": {
      const panes = state.panes.map((pane) =>
        pane.id === action.paneId ? { ...pane, activeTabIndex: action.tabIndex } : pane,
      );
      return { ...state, panes, focusedPaneId: action.paneId };
    }

    case "RESIZE_PANES": {
      const panes = state.panes.map((pane) => {
        if (pane.id === action.leftPaneId) return { ...pane, size: action.leftSize };
        if (pane.id === action.rightPaneId) return { ...pane, size: action.rightSize };
        return pane;
      });
      return { ...state, panes };
    }

    default:
      return state;
  }
}

interface WorkspaceContextValue {
  files: WorkspaceFile[];
  panes: Pane[];
  focusedPaneId: string;
  createFile: (name: string, kind: FileKind) => void;
  renameFile: (id: string, name: string) => void;
  deleteFile: (id: string) => void;
  updateFileContent: (id: string, content: string) => void;
  openFile: (fileId: string, paneId: string) => void;
  closeTab: (paneId: string, tabIndex: number) => void;
  addPane: () => void;
  removePane: (paneId: string) => void;
  setFocusedPane: (paneId: string) => void;
  setActiveTab: (paneId: string, tabIndex: number) => void;
  resizePanes: (leftPaneId: string, rightPaneId: string, leftSize: number, rightSize: number) => void;
}

const WorkspaceContext = createContext<WorkspaceContextValue | null>(null);

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, undefined, initialState);

  const createFile = useCallback(
    (name: string, kind: FileKind) => dispatch({ type: "CREATE_FILE", name, kind }),
    [],
  );
  const renameFile = useCallback(
    (id: string, name: string) => dispatch({ type: "RENAME_FILE", id, name }),
    [],
  );
  const deleteFile = useCallback((id: string) => dispatch({ type: "DELETE_FILE", id }), []);
  const updateFileContent = useCallback(
    (id: string, content: string) => dispatch({ type: "UPDATE_FILE_CONTENT", id, content }),
    [],
  );
  const openFile = useCallback(
    (fileId: string, paneId: string) => dispatch({ type: "OPEN_FILE", fileId, paneId }),
    [],
  );
  const closeTab = useCallback(
    (paneId: string, tabIndex: number) => dispatch({ type: "CLOSE_TAB", paneId, tabIndex }),
    [],
  );
  const addPane = useCallback(() => dispatch({ type: "ADD_PANE" }), []);
  const removePane = useCallback((paneId: string) => dispatch({ type: "REMOVE_PANE", paneId }), []);
  const setFocusedPane = useCallback(
    (paneId: string) => dispatch({ type: "SET_FOCUSED_PANE", paneId }),
    [],
  );
  const setActiveTab = useCallback(
    (paneId: string, tabIndex: number) => dispatch({ type: "SET_ACTIVE_TAB", paneId, tabIndex }),
    [],
  );
  const resizePanes = useCallback(
    (leftPaneId: string, rightPaneId: string, leftSize: number, rightSize: number) =>
      dispatch({ type: "RESIZE_PANES", leftPaneId, rightPaneId, leftSize, rightSize }),
    [],
  );

  const value: WorkspaceContextValue = {
    ...state,
    createFile,
    renameFile,
    deleteFile,
    updateFileContent,
    openFile,
    closeTab,
    addPane,
    removePane,
    setFocusedPane,
    setActiveTab,
    resizePanes,
  };

  return <WorkspaceContext.Provider value={value}>{children}</WorkspaceContext.Provider>;
}

export function useWorkspace(): WorkspaceContextValue {
  const context = useContext(WorkspaceContext);
  if (!context) throw new Error("useWorkspace must be used within a WorkspaceProvider");
  return context;
}

export const MAX_PANE_COUNT = MAX_PANES;
