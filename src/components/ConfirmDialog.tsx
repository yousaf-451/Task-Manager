import { useState } from "react";
import { Modal } from "./Modal";

interface ConfirmDialogProps {
  title: string;
  message: string;
  confirmLabel?: string;
  onConfirm: () => Promise<void>;
  onCancel: () => void;
}

/** Generic "are you sure?" dialog used for destructive actions like delete. */
export function ConfirmDialog({
  title,
  message,
  confirmLabel = "Delete",
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  const [submitting, setSubmitting] = useState(false);

  async function handleConfirm() {
    setSubmitting(true);
    try {
      await onConfirm();
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal title={title} onClose={onCancel}>
      <p style={{ color: "var(--color-ink-soft)", margin: 0 }}>{message}</p>
      <div className="modal__actions">
        <button type="button" className="btn btn-ghost" onClick={onCancel} disabled={submitting}>
          Cancel
        </button>
        <button type="button" className="btn btn-danger" onClick={handleConfirm} disabled={submitting}>
          {submitting ? "Deleting…" : confirmLabel}
        </button>
      </div>
    </Modal>
  );
}
