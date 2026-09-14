import { useMemo, useRef } from "react";
import { useWorkspace } from "../state/WorkspaceContext";
import type { WorkspaceFile } from "../types";
import { highlightJsonLine, highlightOwlLine, type Token } from "../utils/highlight";

const HIGHLIGHTERS: Partial<Record<WorkspaceFile["kind"], (line: string) => Token[]>> = {
  owl: highlightOwlLine,
  json: highlightJsonLine,
};

export default function CodeEditor({ file }: { file: WorkspaceFile }) {
  const { updateFileContent } = useWorkspace();
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const backdropRef = useRef<HTMLDivElement>(null);
  const gutterRef = useRef<HTMLDivElement>(null);

  const lines = useMemo(() => file.content.split("\n"), [file.content]);
  const highlightLine = HIGHLIGHTERS[file.kind];

  function syncScroll() {
    const ta = textareaRef.current;
    if (!ta) return;
    if (backdropRef.current) {
      backdropRef.current.scrollTop = ta.scrollTop;
      backdropRef.current.scrollLeft = ta.scrollLeft;
    }
    if (gutterRef.current) {
      gutterRef.current.scrollTop = ta.scrollTop;
    }
  }

  return (
    <div className="code-editor">
      <div className="code-gutter" ref={gutterRef}>
        {lines.map((_, index) => (
          <div className="code-gutter-line" key={index}>
            {index + 1}
          </div>
        ))}
      </div>
      <div className="code-editor-scroll">
        <div className="code-backdrop" ref={backdropRef} aria-hidden="true">
          {lines.map((line, index) => (
            <div className="code-line" key={index}>
              {line.length === 0 ? " " : renderLine(line, highlightLine)}
            </div>
          ))}
        </div>
        <textarea
          ref={textareaRef}
          className="code-editor-textarea"
          spellCheck={false}
          wrap="off"
          value={file.content}
          onChange={(e) => updateFileContent(file.id, e.target.value)}
          onScroll={syncScroll}
        />
      </div>
    </div>
  );
}

function renderLine(line: string, highlightLine?: (line: string) => Token[]) {
  if (!highlightLine) return line;
  return highlightLine(line).map((token, i) => (
    <span key={i} className={token.className ?? undefined}>
      {token.text}
    </span>
  ));
}
