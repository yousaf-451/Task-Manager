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
  userId: number; // which user this task belongs to (referential integrity: FK -> users.id)
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
  page?: number;
  pageSize?: number;
}

// Pagination metadata returned alongside every task list, so the UI always
// knows the current page, page size, total row count, and total pages -
// without ever fetching more than one page's worth of rows at a time.
export interface Pagination {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

// The full shape of GET /api/tasks's `data` field.
export interface PaginatedTasks {
  tasks: Task[];
  pagination: Pagination;
}

// The signed-in account. Every request is scoped to this user on the
// backend via the session cookie (see api/client.ts), so there is no
// userId to pass around on the frontend anymore.
export interface User {
  id: number;
  name: string;
  email: string;
  createdAt: string;
}

// Body for POST /api/auth/signup.
export interface SignupInput {
  name: string;
  email: string;
  password: string;
}

// Body for POST /api/auth/login.
export interface LoginInput {
  email: string;
  password: string;
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
