import type { TaskStatus } from "../types/task";
import { STATUS_LABELS } from "../types/task";

interface StatusBadgeProps {
  status: TaskStatus;
}

/** Small pill showing a task's status with a colored dot ("ledger stamp"). */
export function StatusBadge({ status }: StatusBadgeProps) {
  return (
    <span className={`status-badge status-badge--${status}`}>
      <span className="status-badge__dot" aria-hidden="true" />
      {STATUS_LABELS[status]}
    </span>
  );
}
