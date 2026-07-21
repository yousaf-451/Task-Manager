import { TaskItem } from "./TaskItem";
import { EmptyState } from "./EmptyState";
import { SkeletonList } from "./SkeletonList";
import type { Task } from "../types/task";

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
  );
}
