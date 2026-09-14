import { WorkspaceProvider, useWorkspace } from "./state/WorkspaceContext";
import TopBar from "./components/TopBar";
import Sidebar from "./components/Sidebar";
import EditorPane from "./components/EditorPane";

function Workbench() {
  const { panes } = useWorkspace();
  return (
    <div className="workbench">
      <Sidebar />
      <div className="pane-row">
        {panes.map((pane) => (
          <EditorPane key={pane.id} pane={pane} />
        ))}
      </div>
    </div>
  );
}

export default function App() {
  return (
    <WorkspaceProvider>
      <div className="app">
        <TopBar />
        <Workbench />
      </div>
    </WorkspaceProvider>
  );
}
