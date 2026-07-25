// Mirrors the Go backend's Status enum (internal/models/task.go).
export type TaskStatus = "pending" | "in_progress" | "completed";

export const TASK_STATUSES: TaskStatus[] = ["pending", "in_progress", "completed"];

export const STATUS_LABELS: Record<TaskStatus, string> = {
  pending: "Pending",
  in_progress: "In Progress",
  completed: "Completed",
};

// Mirrors the Go backend's Priority enum.
export type TaskPriority = "low" | "medium" | "high";

export const TASK_PRIORITIES: TaskPriority[] = ["low", "medium", "high"];

export const PRIORITY_LABELS: Record<TaskPriority, string> = {
  low: "Low",
  medium: "Medium",
  high: "High",
};

// A small, curated accent palette a task can be tagged with. Stored as a
// hex string on the backend so new colors can be added here without a
// migration.
export const TASK_COLORS = [
  "#0e6b5c", // accent green (default)
  "#c98a1a", // amber
  "#2f6fed", // blue
  "#b3432f", // rust
  "#6b7280", // slate
  "#7c3aed", // violet
] as const;

// The shape returned by the API for a single task.
export interface Task {
  id: number;
  title: string;
  description: string;
  dueDate: string; // YYYY-MM-DD
  status: TaskStatus;
  priority: TaskPriority;
  category: string;
  color: string;
  favorite: boolean;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
}

// Payload sent when creating a task (no id / timestamps yet).
export interface CreateTaskInput {
  title: string;
  description: string;
  dueDate: string;
  status: TaskStatus;
  priority: TaskPriority;
  category: string;
  color: string;
}

// Payload sent when updating a task (full replace).
export type UpdateTaskInput = CreateTaskInput;

// Generic envelope every API response follows (see backend handler/response.go).
export interface ApiEnvelope<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export type SortOption = "due_date_asc" | "due_date_desc" | "created_at_desc" | "priority";

export const SORT_LABELS: Record<SortOption, string> = {
  due_date_asc: "Due date: soonest first",
  due_date_desc: "Due date: latest first",
  created_at_desc: "Recently created",
  priority: "Priority: highest first",
};

export interface TaskListParams {
  search?: string;
  status?: TaskStatus | "";
  priority?: TaskPriority | "";
  category?: string;
  favorite?: boolean;
  sortBy?: SortOption;
}

// Dashboard aggregate counts, returned by GET /api/tasks/stats.
export interface TaskStats {
  total: number;
  completed: number;
  pending: number;
  inProgress: number;
  highPriority: number;
  overdue: number;
  upcomingWeek: number;
  favorites: number;
  archived: number;
}
