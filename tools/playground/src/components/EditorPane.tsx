import { useWorkspace } from "../state/WorkspaceContext";
import type { Pane } from "../types";
import TabBar from "./TabBar";
import CodeEditor from "./CodeEditor";
import MarkdownPreview from "./MarkdownPreview";

export default function EditorPane({ pane }: { pane: Pane }) {
  const { files, focusedPaneId, setFocusedPane } = useWorkspace();
  const activeTab = pane.tabs[pane.activeTabIndex];
  const activeFile = activeTab ? files.find((f) => f.id === activeTab.fileId) : undefined;

  return (
    <section
      className={`editor-pane${pane.id === focusedPaneId ? " editor-pane-focused" : ""}`}
      onMouseDown={() => setFocusedPane(pane.id)}
    >
      <TabBar pane={pane} />
      <div className="editor-pane-content">
        {!activeFile && <div className="editor-pane-empty">Select a file to open it here.</div>}
        {activeFile && activeTab?.view === "preview" && <MarkdownPreview file={activeFile} />}
        {activeFile && activeTab?.view === "source" && <CodeEditor file={activeFile} />}
      </div>
    </section>
  );
}
