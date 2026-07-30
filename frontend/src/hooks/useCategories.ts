import { useEffect, useState } from "react";
import { taskApi } from "../api/taskApi";

/** Owns the distinct list of categories currently in use (GET /api/categories).
 * Gated by `enabled` for the same reason as useTasks/useStats. */
export function useCategories(refreshKey: unknown, options?: { enabled?: boolean }): string[] {
  const enabled = options?.enabled ?? true;
  const [categories, setCategories] = useState<string[]>([]);

  useEffect(() => {
    if (!enabled) return;
    let cancelled = false;
    taskApi
      .categories()
      .then((data) => {
        if (!cancelled) setCategories(data);
      })
      .catch(() => {
        // Categories are a "nice to have" filter aid; fail silently rather
        // than surfacing a toast for a non-critical dropdown.
      });
    return () => {
      cancelled = true;
    };
  }, [refreshKey, enabled]);

  return categories;
}
