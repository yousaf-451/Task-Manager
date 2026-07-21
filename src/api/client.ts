import type { ApiEnvelope } from "../types/task";
import { API_REQUEST_TIMEOUT_MS } from "../constants";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api";

/** Thrown whenever the API responds with success: false, or a non-2xx status. */
export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

/**
 * Thin wrapper around fetch() that:
 *  - prefixes the API base URL
 *  - always sends/expects JSON
 *  - unwraps the {success, data, error} envelope
 *  - aborts and throws a friendly ApiError if the server doesn't respond
 *    within API_REQUEST_TIMEOUT_MS, so the UI never hangs forever
 *  - throws ApiError on failure, so callers can use try/catch uniformly
 */
async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), API_REQUEST_TIMEOUT_MS);

  let response: Response;

  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      signal: controller.signal,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") {
      throw new ApiError("The request took too long and was cancelled. Please try again.", 0);
    }
    throw new ApiError("Could not reach the server. Check your connection and try again.", 0);
  } finally {
    clearTimeout(timeoutId);
  }

  // 204 No Content (not currently used, but handled defensively).
  if (response.status === 204) {
    return undefined as T;
  }

  let body: ApiEnvelope<T> | undefined;
  try {
    body = await response.json();
  } catch {
    // Non-JSON response (e.g. the API is down and a proxy returned HTML).
  }

  if (!response.ok || !body?.success) {
    const message = body?.error || `Request failed with status ${response.status}`;
    throw new ApiError(message, response.status);
  }

  return body.data as T;
}

export const apiClient = {
  get: <T>(path: string) => request<T>(path, { method: "GET" }),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "POST", body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(body) }),
  patch: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PATCH", body: JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
};
