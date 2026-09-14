import { Fragment, useRef } from "react";
import { WorkspaceProvider, useWorkspace } from "./state/WorkspaceContext";
import TopBar from "./components/TopBar";
import Sidebar from "./components/Sidebar";
import EditorPane from "./components/EditorPane";
import PaneResizer from "./components/PaneResizer";

function Workbench() {
  const { panes } = useWorkspace();
  const paneRowRef = useRef<HTMLDivElement>(null);
  const gridTemplateColumns = panes.map((pane) => `${pane.size}fr`).join(" 6px ");

  return (
    <div className="workbench">
      <Sidebar />
      <div className="pane-row" ref={paneRowRef} style={{ gridTemplateColumns }}>
        {panes.map((pane, index) => (
          <Fragment key={pane.id}>
            {index > 0 && (
              <PaneResizer
                leftPane={panes[index - 1]}
                rightPane={pane}
                paneCount={panes.length}
                containerRef={paneRowRef}
              />
            )}
            <EditorPane pane={pane} />
          </Fragment>
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
