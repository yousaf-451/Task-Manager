import { StatusBadge } from "./StatusBadge";
import { PriorityBadge } from "./PriorityBadge";
import { CategoryTag } from "./CategoryTag";
import type { Task } from "../types/task";

interface TaskItemProps {
  task: Task;
  selectable: boolean;
  selected: boolean;
  onToggleSelect: (task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
  onDuplicate: (task: Task) => void;
  onToggleFavorite: (task: Task) => void;
  onArchive: (task: Task) => void;
}

function formatDueDate(dueDate: string): string {
  // dueDate is "YYYY-MM-DD"; parse as local date to avoid timezone shifting.
  const [year, month, day] = dueDate.split("-").map(Number);
  const date = new Date(year, month - 1, day);
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}

function isOverdue(dueDate: string, status: Task["status"]): boolean {
  if (status === "completed") return false;
  const [year, month, day] = dueDate.split("-").map(Number);
  const due = new Date(year, month - 1, day);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return due < today;
}

export function TaskItem({
  task,
  selectable,
  selected,
  onToggleSelect,
  onEdit,
  onDelete,
  onDuplicate,
  onToggleFavorite,
  onArchive,
}: TaskItemProps) {
  const overdue = isOverdue(task.dueDate, task.status);

  return (
    <article className={`task-card ${selected ? "task-card--selected" : ""}`}>
      <span className="task-card__accent" style={{ background: task.color }} aria-hidden="true" />

      {selectable && (
        <input
          type="checkbox"
          className="task-card__checkbox"
          checked={selected}
          onChange={() => onToggleSelect(task)}
          aria-label={`Select ${task.title}`}
        />
      )}

      <div className="task-card__main">
        <div className="task-card__title-row">
          <h3 className="task-card__title">{task.title}</h3>
          <button
            className={`favorite-star ${task.favorite ? "favorite-star--on" : ""}`}
            onClick={() => onToggleFavorite(task)}
            aria-label={task.favorite ? `Remove ${task.title} from favorites` : `Favorite ${task.title}`}
            aria-pressed={task.favorite}
            type="button"
          >
            ★
          </button>
        </div>
        {task.description && <p className="task-card__description">{task.description}</p>}
        <div className="task-card__meta">
          <StatusBadge status={task.status} />
          <PriorityBadge priority={task.priority} />
          <CategoryTag category={task.category} />
          <span className={`task-card__due ${overdue ? "task-card__due--overdue" : ""}`}>
            {overdue ? "Overdue · " : "Due "}
            {formatDueDate(task.dueDate)}
          </span>
        </div>
      </div>

      <div className="task-card__actions">
        <button className="btn btn-ghost btn-sm" onClick={() => onEdit(task)} aria-label={`Edit ${task.title}`}>
          Edit
        </button>
        <button
          className="btn btn-ghost btn-sm"
          onClick={() => onDuplicate(task)}
          aria-label={`Duplicate ${task.title}`}
        >
          Duplicate
        </button>
        <button
          className="btn btn-ghost btn-sm"
          onClick={() => onArchive(task)}
          aria-label={`Archive ${task.title}`}
        >
          Archive
        </button>
        <button
          className="btn btn-ghost btn-sm"
          onClick={() => onDelete(task)}
          aria-label={`Delete ${task.title}`}
        >
          Delete
        </button>
      </div>
    </article>
  );
}
