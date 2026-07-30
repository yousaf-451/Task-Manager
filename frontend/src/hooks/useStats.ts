import { useCallback, useEffect, useState } from "react";
import { taskApi } from "../api/taskApi";
import { ApiError } from "../api/client";
import type { TaskStats } from "../types/task";

interface UseStatsResult {
  stats: TaskStats | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

/** Owns the dashboard's aggregate counts (GET /api/tasks/stats). Gated by
 * `enabled` for the same reason as useTasks - no fetch before login. */
export function useStats(options?: { enabled?: boolean }): UseStatsResult {
  const enabled = options?.enabled ?? true;
  const [stats, setStats] = useState<TaskStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = useCallback(async () => {
    if (!enabled) return;
    setLoading(true);
    setError(null);
    try {
      const data = await taskApi.stats();
      setStats(data);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to load dashboard stats.");
    } finally {
      setLoading(false);
    }
  }, [enabled]);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  return { stats, loading, error, refetch: fetchStats };
}
