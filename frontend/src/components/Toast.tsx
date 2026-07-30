export interface ToastMessage {
  id: number;
  kind: "success" | "error";
  text: string;
}

interface ToastStackProps {
  toasts: ToastMessage[];
}

/** Fixed-position stack of success/error notifications, bottom-right. */
export function ToastStack({ toasts }: ToastStackProps) {
  if (toasts.length === 0) return null;

  return (
    <div className="toast-stack" role="status" aria-live="polite">
      {toasts.map((t) => (
        <div key={t.id} className={`toast toast--${t.kind}`}>
          {t.text}
        </div>
      ))}
    </div>
  );
}
