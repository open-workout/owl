import { useEffect } from "react";

export default function ConfirmDialog({
  title,
  message,
  confirmLabel = "Confirm",
  onConfirm,
  onCancel,
}: {
  title: string;
  message: string;
  confirmLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") onCancel();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onCancel]);

  return (
    <div className="modal-overlay" onMouseDown={onCancel}>
      <div className="modal" role="alertdialog" aria-modal="true" onMouseDown={(e) => e.stopPropagation()}>
        <div className="modal-title">{title}</div>
        <div className="modal-message">{message}</div>
        <div className="modal-actions">
          <button type="button" className="modal-button" onClick={onCancel}>
            Cancel
          </button>
          <button type="button" className="modal-button modal-button-danger" onClick={onConfirm} autoFocus>
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
