import { apiClient } from "./client";
import type {
  CreateTaskInput,
  PaginatedTasks,
  Task,
  TaskListParams,
  TaskStats,
  UpdateTaskInput,
} from "../types/task";

function buildQuery(params: TaskListParams): string {
  const search = new URLSearchParams();
  if (params.search) search.set("search", params.search);
  if (params.status) search.set("status", params.status);
  if (params.priority) search.set("priority", params.priority);
  if (params.category) search.set("category", params.category);
  if (params.favorite) search.set("favorite", "true");
  if (params.sortBy) search.set("sortBy", params.sortBy);
  if (params.page) search.set("page", String(params.page));
  if (params.pageSize) search.set("pageSize", String(params.pageSize));
  const qs = search.toString();
  return qs ? `?${qs}` : "";
}

export const taskApi = {
  // Returns one page of tasks plus pagination metadata (total, totalPages,
  // etc). The server never sends back more than `pageSize` rows, no
  // matter how many tasks the user has in total.
  list: (params: TaskListParams = {}) => apiClient.get<PaginatedTasks>(`/tasks${buildQuery(params)}`),

  getById: (id: number) => apiClient.get<Task>(`/tasks/${id}`),

  create: (input: CreateTaskInput) => apiClient.post<Task>("/tasks", input),

  update: (id: number, input: UpdateTaskInput) => apiClient.put<Task>(`/tasks/${id}`, input),

  remove: (id: number) => apiClient.delete<{ id: number }>(`/tasks/${id}`),

  duplicate: (id: number) => apiClient.post<Task>(`/tasks/${id}/duplicate`, {}),

  setFavorite: (id: number, favorite: boolean) =>
    apiClient.patch<Task>(`/tasks/${id}/favorite`, { favorite }),

  setArchived: (id: number, archived: boolean) =>
    apiClient.patch<Task>(`/tasks/${id}/archive`, { archived }),

  bulkDelete: (ids: number[]) => apiClient.post<{ deleted: number }>("/tasks/bulk/delete", { ids }),

  bulkComplete: (ids: number[]) =>
    apiClient.post<{ completed: number }>("/tasks/bulk/complete", { ids }),

  stats: () => apiClient.get<TaskStats>("/tasks/stats"),

  categories: () => apiClient.get<string[]>("/categories"),
};
