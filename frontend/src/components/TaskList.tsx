import { TaskItem } from "./TaskItem";
import { EmptyState } from "./EmptyState";
import { SkeletonList } from "./SkeletonList";
import type { Pagination, Task } from "../types/task";

interface TaskListProps {
  tasks: Task[];
  loading: boolean;
  hasFilters: boolean;
  selectable: boolean;
  selectedIds: Set<number>;
  onToggleSelect: (task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
  onDuplicate: (task: Task) => void;
  onToggleFavorite: (task: Task) => void;
  onArchive: (task: Task) => void;
  onAddTask: () => void;
  pagination: Pagination;
  onPageChange: (page: number) => void;
}

export function TaskList({
  tasks,
  loading,
  hasFilters,
  selectable,
  selectedIds,
  onToggleSelect,
  onEdit,
  onDelete,
  onDuplicate,
  onToggleFavorite,
  onArchive,
  onAddTask,
  pagination,
  onPageChange,
}: TaskListProps) {
  if (loading) {
    return (
      <>
        <span className="sr-only" role="status" aria-live="polite">
          Loading tasks…
        </span>
        <SkeletonList />
      </>
    );
  }

  if (tasks.length === 0) {
    return <EmptyState hasFilters={hasFilters} onAddTask={onAddTask} />;
  }

  return (
    <>
      <div className="task-list">
        {tasks.map((task) => (
          <TaskItem
            key={task.id}
            task={task}
            selectable={selectable}
            selected={selectedIds.has(task.id)}
            onToggleSelect={onToggleSelect}
            onEdit={onEdit}
            onDelete={onDelete}
            onDuplicate={onDuplicate}
            onToggleFavorite={onToggleFavorite}
            onArchive={onArchive}
          />
        ))}
      </div>

      {/*
        Server-side pagination: the list above only ever contains one
        page's worth of rows (see repository.TaskFilter's LIMIT/OFFSET on
        the backend). This is what keeps the app fast even if the table
        grows to millions of rows - the browser never has to download or
        render more than `pageSize` tasks at a time.
      */}
      {pagination.totalPages > 1 && (
        <nav className="task-pagination" aria-label="Task list pages">
          <button
            type="button"
            className="btn btn-ghost"
            onClick={() => onPageChange(pagination.page - 1)}
            disabled={pagination.page <= 1}
          >
            ← Prev
          </button>
          <span className="task-pagination__label">
            Page {pagination.page} of {pagination.totalPages} · {pagination.total} task
            {pagination.total === 1 ? "" : "s"} total
          </span>
          <button
            type="button"
            className="btn btn-ghost"
            onClick={() => onPageChange(pagination.page + 1)}
            disabled={pagination.page >= pagination.totalPages}
          >
            Next →
          </button>
        </nav>
      )}
    </>
  );
}
