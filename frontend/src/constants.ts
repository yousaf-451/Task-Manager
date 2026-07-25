/**
 * Centralized, app-wide constants. Keeping these in one place makes them
 * easy to tune (and easy to find) instead of hunting for magic numbers
 * scattered across components.
 */

/** How long to wait after the user stops typing before firing a search request. */
export const SEARCH_DEBOUNCE_MS = 350;

/** How long a toast notification stays on screen before auto-dismissing. */
export const TOAST_DURATION_MS = 3200;

/** Client-side timeout for any single API request (see api/client.ts). */
export const API_REQUEST_TIMEOUT_MS = 10_000;

/** Default title length limit, mirrored from the backend's validation rule. */
export const TASK_TITLE_MAX_LENGTH = 150;

/** Default description length limit, mirrored from the backend's schema. */
export const TASK_DESCRIPTION_MAX_LENGTH = 2000;
