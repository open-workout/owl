import { useWorkspace } from "../state/WorkspaceContext";
import type { WorkspaceFile } from "../types";

export default function CodeEditor({ file }: { file: WorkspaceFile }) {
  const { updateFileContent } = useWorkspace();
  return (
    <textarea
      className="code-editor"
      spellCheck={false}
      value={file.content}
      onChange={(e) => updateFileContent(file.id, e.target.value)}
    />
  );
}
