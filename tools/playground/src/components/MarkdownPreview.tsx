import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { WorkspaceFile } from "../types";

export default function MarkdownPreview({ file }: { file: WorkspaceFile }) {
  return (
    <div className="markdown-preview">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{file.content}</ReactMarkdown>
    </div>
  );
}
