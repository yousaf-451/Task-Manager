import { forwardRef } from "react";
import type { SortOption, TaskPriority, TaskStatus } from "../types/task";
import { PRIORITY_LABELS, SORT_LABELS, STATUS_LABELS, TASK_PRIORITIES, TASK_STATUSES } from "../types/task";

interface SearchFilterBarProps {
  search: string;
  onSearchChange: (value: string) => void;
  status: TaskStatus | "";
  onStatusChange: (value: TaskStatus | "") => void;
  priority: TaskPriority | "";
  onPriorityChange: (value: TaskPriority | "") => void;
  category: string;
  onCategoryChange: (value: string) => void;
  categories: string[];
  favoriteOnly: boolean;
  onFavoriteOnlyChange: (value: boolean) => void;
  sortBy: SortOption;
  onSortByChange: (value: SortOption) => void;
  selectMode: boolean;
  onToggleSelectMode: () => void;
}

/** Toolbar above the task list: free-text search, filters, sort, and bulk-select toggle. */
export const SearchFilterBar = forwardRef<HTMLInputElement, SearchFilterBarProps>(function SearchFilterBar(
  {
    search,
    onSearchChange,
    status,
    onStatusChange,
    priority,
    onPriorityChange,
    category,
    onCategoryChange,
    categories,
    favoriteOnly,
    onFavoriteOnlyChange,
    sortBy,
    onSortByChange,
    selectMode,
    onToggleSelectMode,
  },
  searchRef
) {
  return (
    <div className="toolbar">
      <input
        ref={searchRef}
        type="text"
        placeholder="Search by title or description… (press /)"
        value={search}
        onChange={(e) => onSearchChange(e.target.value)}
        aria-label="Search tasks"
      />

      <select
        value={status}
        onChange={(e) => onStatusChange(e.target.value as TaskStatus | "")}
        aria-label="Filter by status"
      >
        <option value="">All statuses</option>
        {TASK_STATUSES.map((s) => (
          <option key={s} value={s}>
            {STATUS_LABELS[s]}
          </option>
        ))}
      </select>

      <select
        value={priority}
        onChange={(e) => onPriorityChange(e.target.value as TaskPriority | "")}
        aria-label="Filter by priority"
      >
        <option value="">All priorities</option>
        {TASK_PRIORITIES.map((p) => (
          <option key={p} value={p}>
            {PRIORITY_LABELS[p]}
          </option>
        ))}
      </select>

      {categories.length > 0 && (
        <select
          value={category}
          onChange={(e) => onCategoryChange(e.target.value)}
          aria-label="Filter by category"
        >
          <option value="">All categories</option>
          {categories.map((c) => (
            <option key={c} value={c}>
              {c}
            </option>
          ))}
        </select>
      )}

      <select
        value={sortBy}
        onChange={(e) => onSortByChange(e.target.value as SortOption)}
        aria-label="Sort tasks"
      >
        {(Object.keys(SORT_LABELS) as SortOption[]).map((option) => (
          <option key={option} value={option}>
            {SORT_LABELS[option]}
          </option>
        ))}
      </select>

      <label className="toolbar__checkbox">
        <input
          type="checkbox"
          checked={favoriteOnly}
          onChange={(e) => onFavoriteOnlyChange(e.target.checked)}
        />
        Favorites only
      </label>

      <button
        type="button"
        className={`btn btn-ghost btn-sm ${selectMode ? "btn-ghost--active" : ""}`}
        onClick={onToggleSelectMode}
      >
        {selectMode ? "Cancel select" : "Select"}
      </button>
    </div>
  );
});
