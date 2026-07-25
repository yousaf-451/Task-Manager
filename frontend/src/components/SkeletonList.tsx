interface SkeletonListProps {
  /** Number of placeholder cards to render. */
  count?: number;
}

/**
 * Renders a handful of pulsing placeholder cards shaped like TaskItem, so
 * the layout doesn't jump once real data arrives. Purely decorative
 * (aria-hidden) — the "Loading tasks…" text is announced separately for
 * screen readers by the live region in TaskList.
 */
export function SkeletonList({ count = 4 }: SkeletonListProps) {
  return (
    <div className="task-list" aria-hidden="true">
      {Array.from({ length: count }).map((_, i) => (
        <div className="task-card task-card--skeleton" key={i}>
          <div className="task-card__main">
            <div className="skeleton-block skeleton-block--title" />
            <div className="skeleton-block skeleton-block--text" />
            <div className="task-card__meta">
              <div className="skeleton-block skeleton-block--badge" />
              <div className="skeleton-block skeleton-block--date" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
