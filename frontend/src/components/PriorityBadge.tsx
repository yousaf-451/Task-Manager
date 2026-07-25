import type { TaskPriority } from "../types/task";
import { PRIORITY_LABELS } from "../types/task";

interface PriorityBadgeProps {
  priority: TaskPriority;
}

/** Small pill showing a task's priority, styled to match StatusBadge. */
export function PriorityBadge({ priority }: PriorityBadgeProps) {
  return (
    <span className={`priority-badge priority-badge--${priority}`}>
      {priority === "high" && (
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none" aria-hidden="true">
          <path d="M5 1.2L8.5 8.8H1.5L5 1.2Z" fill="currentColor" />
        </svg>
      )}
      {PRIORITY_LABELS[priority]}
    </span>
  );
}
