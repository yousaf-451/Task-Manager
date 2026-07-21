import { useEffect, useState } from "react";
import { taskApi } from "../api/taskApi";

/** Owns the distinct list of categories currently in use (GET /api/categories). */
export function useCategories(refreshKey: unknown): string[] {
  const [categories, setCategories] = useState<string[]>([]);

  useEffect(() => {
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
  }, [refreshKey]);

  return categories;
}
