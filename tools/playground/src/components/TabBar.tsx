import { useWorkspace, MAX_PANE_COUNT } from "../state/WorkspaceContext";
import type { Pane, WorkspaceFile } from "../types";

function tabLabel(file: WorkspaceFile, view: "source" | "preview"): string {
  return view === "preview" ? `${file.name} (Preview)` : file.name;
}

export default function TabBar({ pane }: { pane: Pane }) {
  const { files, panes, closeTab, setActiveTab, addPane, removePane } = useWorkspace();

  return (
    <div className="tab-bar">
      <div className="tab-list">
        {pane.tabs.map((tab, index) => {
          const file = files.find((f) => f.id === tab.fileId);
          if (!file) return null;
          return (
            <div
              key={`${tab.fileId}-${tab.view}`}
              className={`tab${index === pane.activeTabIndex ? " tab-active" : ""}`}
              onClick={() => setActiveTab(pane.id, index)}
            >
              <span className="tab-label">{tabLabel(file, tab.view)}</span>
              <button
                type="button"
                className="tab-close"
                onClick={(e) => {
                  e.stopPropagation();
                  closeTab(pane.id, index);
                }}
                aria-label={`Close ${tabLabel(file, tab.view)}`}
              >
                ×
              </button>
            </div>
          );
        })}
      </div>
      <div className="pane-controls">
        {panes.length < MAX_PANE_COUNT && (
          <button type="button" onClick={addPane} title="Split into a new pane">
            ⊞ Split
          </button>
        )}
        {panes.length > 1 && (
          <button type="button" onClick={() => removePane(pane.id)} title="Close this pane">
            Close Pane
          </button>
        )}
      </div>
    </div>
  );
}
