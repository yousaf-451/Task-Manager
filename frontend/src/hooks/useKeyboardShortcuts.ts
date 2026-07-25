import { useEffect } from "react";

interface ShortcutMap {
  /** Focus the search input. */
  onSearch?: () => void;
  /** Open the "add task" form. */
  onNew?: () => void;
}

function isTypingTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false;
  const tag = target.tagName;
  return tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT" || target.isContentEditable;
}

/**
 * Registers a small set of global keyboard shortcuts:
 *  - "/"  focuses the search box
 *  - "n"  opens the add-task form
 * Both are ignored while the user is already typing in a field, and while
 * a modifier key (which usually means a browser/OS shortcut) is held.
 */
export function useKeyboardShortcuts({ onSearch, onNew }: ShortcutMap) {
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (isTypingTarget(e.target)) return;

      if (e.key === "/") {
        e.preventDefault();
        onSearch?.();
      } else if (e.key === "n" || e.key === "N") {
        e.preventDefault();
        onNew?.();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [onSearch, onNew]);
}
