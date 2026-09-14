import { useState } from "react";
import { useWorkspace } from "../state/WorkspaceContext";
import { FILE_KIND_ICON, kindFromFileName } from "../utils/fileTypes";
import ConfirmDialog from "./ConfirmDialog";
import type { WorkspaceFile } from "../types";

export default function Sidebar() {
  const { files, focusedPaneId, createFile, renameFile, deleteFile, openFile } = useWorkspace();
  const [newFileName, setNewFileName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [renamingId, setRenamingId] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState("");
  const [pendingDelete, setPendingDelete] = useState<WorkspaceFile | null>(null);

  function handleCreate(event: React.FormEvent) {
    event.preventDefault();
    const name = newFileName.trim();
    const kind = kindFromFileName(name);
    if (!kind) {
      setError("Name must end in .owl, .json, or .md");
      return;
    }
    createFile(name, kind);
    setNewFileName("");
    setError(null);
  }

  function startRename(id: string, currentName: string) {
    setRenamingId(id);
    setRenameValue(currentName);
  }

  function commitRename(id: string) {
    const name = renameValue.trim();
    if (name && kindFromFileName(name)) {
      renameFile(id, name);
    }
    setRenamingId(null);
  }

  function confirmDelete() {
    if (pendingDelete) deleteFile(pendingDelete.id);
    setPendingDelete(null);
  }

  return (
    <aside className="sidebar">
      <div className="sidebar-header">Files</div>
      {files.length === 0 && <div className="sidebar-empty">No files yet</div>}
      <ul className="file-tree">
        {files.map((file) => (
          <li key={file.id} className="file-tree-row">
            {renamingId === file.id ? (
              <input
                autoFocus
                className="file-rename-input"
                value={renameValue}
                onChange={(e) => setRenameValue(e.target.value)}
                onBlur={() => commitRename(file.id)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") commitRename(file.id);
                  if (e.key === "Escape") setRenamingId(null);
                }}
              />
            ) : (
              <button
                type="button"
                className="file-tree-item"
                onClick={() => openFile(file.id, focusedPaneId)}
                onDoubleClick={() => startRename(file.id, file.name)}
                title="Click to open, double-click to rename"
              >
                <span className="file-icon">{FILE_KIND_ICON[file.kind]}</span>
                <span className="file-name">{file.name}</span>
              </button>
            )}
            <button
              type="button"
              className="file-delete"
              onClick={() => setPendingDelete(file)}
              title="Delete file"
              aria-label={`Delete ${file.name}`}
            >
              ×
            </button>
          </li>
        ))}
      </ul>
      <form className="new-file-form" onSubmit={handleCreate}>
        <input
          type="text"
          placeholder="new-file.owl"
          value={newFileName}
          onChange={(e) => {
            setNewFileName(e.target.value);
            setError(null);
          }}
        />
        <button type="submit">+ New File</button>
        {error && <div className="new-file-error">{error}</div>}
      </form>
      {pendingDelete && (
        <ConfirmDialog
          title="Delete file"
          message={`Delete "${pendingDelete.name}"? This can't be undone.`}
          confirmLabel="Delete"
          onConfirm={confirmDelete}
          onCancel={() => setPendingDelete(null)}
        />
      )}
    </aside>
  );
}
