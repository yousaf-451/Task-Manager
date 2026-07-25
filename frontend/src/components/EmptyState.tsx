interface EmptyStateProps {
  hasFilters: boolean;
  onAddTask: () => void;
}

/** Shown when the list is empty - message and CTA differ if filters are active. */
export function EmptyState({ hasFilters, onAddTask }: EmptyStateProps) {
  return (
    <div className="state-panel">
      <p className="state-panel__title">
        {hasFilters ? "No tasks match your filters" : "No tasks yet"}
      </p>
      <p style={{ margin: "0 0 16px" }}>
        {hasFilters
          ? "Try a different search term or clear the status filter."
          : "Add your first task to get the dashboard started."}
      </p>
      {!hasFilters && (
        <button className="btn btn-accent" onClick={onAddTask}>
          + Add task
        </button>
      )}
    </div>
  );
}
