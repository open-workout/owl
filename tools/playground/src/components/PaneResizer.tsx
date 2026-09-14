import type { RefObject } from "react";
import { useWorkspace, MIN_PANE_SIZE } from "../state/WorkspaceContext";
import type { Pane } from "../types";

const RESIZER_WIDTH_PX = 6;

export default function PaneResizer({
  leftPane,
  rightPane,
  paneCount,
  containerRef,
}: {
  leftPane: Pane;
  rightPane: Pane;
  paneCount: number;
  containerRef: RefObject<HTMLDivElement | null>;
}) {
  const { resizePanes } = useWorkspace();

  function onMouseDown(event: React.MouseEvent) {
    event.preventDefault();
    const container = containerRef.current;
    if (!container) return;

    const startX = event.clientX;
    const leftStart = leftPane.size;
    const rightStart = rightPane.size;
    const pairTotal = leftStart + rightStart;
    // Weights sum to 100 across all panes, so the space available to fr
    // tracks (container width minus the fixed-width resizers) maps
    // linearly onto that 0-100 weight scale.
    const fixedResizerWidth = (paneCount - 1) * RESIZER_WIDTH_PX;
    const frTrackWidth = container.getBoundingClientRect().width - fixedResizerWidth;
    const weightPerPixel = 100 / frTrackWidth;

    document.body.style.cursor = "col-resize";

    function handleMouseMove(moveEvent: MouseEvent) {
      const deltaWeight = (moveEvent.clientX - startX) * weightPerPixel;
      const min = Math.min(MIN_PANE_SIZE, pairTotal / 2);
      const newLeft = Math.max(min, Math.min(pairTotal - min, leftStart + deltaWeight));
      resizePanes(leftPane.id, rightPane.id, newLeft, pairTotal - newLeft);
    }

    function handleMouseUp() {
      document.body.style.cursor = "";
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    }

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);
  }

  return <div className="pane-resizer" onMouseDown={onMouseDown} />;
}
