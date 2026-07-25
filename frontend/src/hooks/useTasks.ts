import { useCallback, useEffect, useState } from "react";
import { taskApi } from "../api/taskApi";
import { ApiError } from "../api/client";
import type { CreateTaskInput, Task, TaskListParams, UpdateTaskInput } from "../types/task";

interface UseTasksResult {
  tasks: Task[];
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
  createTask: (input: CreateTaskInput) => Promise<Task>;
  updateTask: (id: number, input: UpdateTaskInput) => Promise<Task>;
  deleteTask: (id: number) => Promise<void>;
  duplicateTask: (id: number) => Promise<Task>;
  toggleFavorite: (task: Task) => Promise<void>;
  archiveTask: (task: Task) => Promise<void>;
  bulkDelete: (ids: number[]) => Promise<void>;
  bulkComplete: (ids: number[]) => Promise<void>;
}

/**
 * Owns the task list for the given filter/sort params, and exposes CRUD
 * helpers that keep local state in sync after each mutation so the UI
 * updates immediately without a full page reload.
 */
export function useTasks(params: TaskListParams): UseTasksResult {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await taskApi.list(params);
      setTasks(data);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to load tasks.");
    } finally {
      setLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params.search, params.status, params.priority, params.category, params.favorite, params.sortBy]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const createTask = useCallback(async (input: CreateTaskInput) => {
    const created = await taskApi.create(input);
    setTasks((prev) => [created, ...prev]);
    return created;
  }, []);

  const updateTask = useCallback(async (id: number, input: UpdateTaskInput) => {
    const updated = await taskApi.update(id, input);
    setTasks((prev) => prev.map((t) => (t.id === id ? updated : t)));
    return updated;
  }, []);

  const deleteTask = useCallback(async (id: number) => {
    await taskApi.remove(id);
    setTasks((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const duplicateTask = useCallback(async (id: number) => {
    const created = await taskApi.duplicate(id);
    setTasks((prev) => [created, ...prev]);
    return created;
  }, []);

  // Optimistic: flip the flag locally right away, then reconcile with the
  // server response (or roll back on failure) so starring a task feels instant.
  const toggleFavorite = useCallback(async (task: Task) => {
    const next = !task.favorite;
    setTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, favorite: next } : t)));
    try {
      const updated = await taskApi.setFavorite(task.id, next);
      setTasks((prev) => prev.map((t) => (t.id === task.id ? updated : t)));
    } catch (err) {
      setTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, favorite: !next } : t)));
      throw err;
    }
  }, []);

  // Archiving removes the task from the current (non-archived) view.
  const archiveTask = useCallback(async (task: Task) => {
    await taskApi.setArchived(task.id, true);
    setTasks((prev) => prev.filter((t) => t.id !== task.id));
  }, []);

  const bulkDelete = useCallback(async (ids: number[]) => {
    await taskApi.bulkDelete(ids);
    setTasks((prev) => prev.filter((t) => !ids.includes(t.id)));
  }, []);

  const bulkComplete = useCallback(async (ids: number[]) => {
    await taskApi.bulkComplete(ids);
    setTasks((prev) => prev.map((t) => (ids.includes(t.id) ? { ...t, status: "completed" } : t)));
  }, []);

  return {
    tasks,
    loading,
    error,
    refetch: fetchTasks,
    createTask,
    updateTask,
    deleteTask,
    duplicateTask,
    toggleFavorite,
    archiveTask,
    bulkDelete,
    bulkComplete,
  };
}
