import { useEffect, useMemo, useRef, useState } from "react";
import { Sidebar } from "./components/Sidebar";
import type { View } from "./components/Sidebar";
import { Header } from "./components/Header";
import { Dashboard } from "./components/Dashboard";
import { SearchFilterBar } from "./components/SearchFilterBar";
import { TaskList } from "./components/TaskList";
import { TaskForm } from "./components/TaskForm";
import { BulkActionBar } from "./components/BulkActionBar";
import { ConfirmDialog } from "./components/ConfirmDialog";
import { ToastStack } from "./components/Toast";
import type { ToastMessage } from "./components/Toast";
import { useTasks } from "./hooks/useTasks";
import { useStats } from "./hooks/useStats";
import { useCategories } from "./hooks/useCategories";
import { useKeyboardShortcuts } from "./hooks/useKeyboardShortcuts";
import { ApiError } from "./api/client";
import { SEARCH_DEBOUNCE_MS, TOAST_DURATION_MS } from "./constants";
import type { CreateTaskInput, SortOption, Task, TaskPriority, TaskStatus } from "./types/task";

export default function App() {
  const [view, setView] = useState<View>("dashboard");
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [status, setStatus] = useState<TaskStatus | "">("");
  const [priority, setPriority] = useState<TaskPriority | "">("");
  const [category, setCategory] = useState("");
  const [favoriteOnly, setFavoriteOnly] = useState(false);
  const [sortBy, setSortBy] = useState<SortOption>("due_date_asc");

  const [selectMode, setSelectMode] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

  const [formOpen, setFormOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | undefined>(undefined);
  const [deletingTask, setDeletingTask] = useState<Task | null>(null);
  const [bulkDeleteConfirm, setBulkDeleteConfirm] = useState(false);

  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const searchInputRef = useRef<HTMLInputElement>(null);

  // Debounce search input so we don't hit the API on every keystroke.
  useEffect(() => {
    const id = setTimeout(() => setDebouncedSearch(search), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(id);
  }, [search]);

  const listParams = useMemo(
    () => ({ search: debouncedSearch, status, priority, category, favorite: favoriteOnly, sortBy }),
    [debouncedSearch, status, priority, category, favoriteOnly, sortBy]
  );

  const {
    tasks,
    loading,
    error,
    createTask,
    updateTask,
    deleteTask,
    duplicateTask,
    toggleFavorite,
    archiveTask,
    bulkDelete,
    bulkComplete,
  } = useTasks(listParams);

  const { stats, loading: statsLoading, error: statsError, refetch: refetchStats } = useStats();
  const categories = useCategories(tasks.length);

  function pushToast(kind: ToastMessage["kind"], text: string) {
    const id = Date.now() + Math.random();
    setToasts((prev) => [...prev, { id, kind, text }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, TOAST_DURATION_MS);
  }

  function openAddForm() {
    setEditingTask(undefined);
    setFormOpen(true);
  }

  function openEditForm(task: Task) {
    setEditingTask(task);
    setFormOpen(true);
  }

  function closeForm() {
    setFormOpen(false);
    setEditingTask(undefined);
  }

  async function handleFormSubmit(input: CreateTaskInput) {
    try {
      if (editingTask) {
        await updateTask(editingTask.id, input);
        pushToast("success", `"${input.title}" was updated.`);
      } else {
        await createTask(input);
        pushToast("success", `"${input.title}" was added.`);
      }
      closeForm();
      refetchStats();
    } catch (err) {
      // Re-throw so the form can show the inline error too; toast gives a
      // second, more visible signal for the common failure cases.
      const message = err instanceof ApiError ? err.message : "Something went wrong.";
      pushToast("error", message);
      throw err;
    }
  }

  async function handleConfirmDelete() {
    if (!deletingTask) return;
    try {
      await deleteTask(deletingTask.id);
      pushToast("success", `"${deletingTask.title}" was deleted.`);
      refetchStats();
    } catch (err) {
      pushToast("error", err instanceof ApiError ? err.message : "Failed to delete task.");
    } finally {
      setDeletingTask(null);
    }
  }

  async function handleDuplicate(task: Task) {
    try {
      await duplicateTask(task.id);
      pushToast("success", `"${task.title}" was duplicated.`);
      refetchStats();
    } catch (err) {
      pushToast("error", err instanceof ApiError ? err.message : "Failed to duplicate task.");
    }
  }

  async function handleToggleFavorite(task: Task) {
    try {
      await toggleFavorite(task);
      refetchStats();
    } catch (err) {
      pushToast("error", err instanceof ApiError ? err.message : "Failed to update favorite.");
    }
  }

  async function handleArchive(task: Task) {
    try {
      await archiveTask(task);
      pushToast("success", `"${task.title}" was archived.`);
      refetchStats();
    } catch (err) {
      pushToast("error", err instanceof ApiError ? err.message : "Failed to archive task.");
    }
  }

  function toggleSelect(task: Task) {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(task.id)) next.delete(task.id);
      else next.add(task.id);
      return next;
    });
  }

  function toggleSelectMode() {
    setSelectMode((prev) => !prev);
    setSelectedIds(new Set());
  }

  async function handleBulkComplete() {
    const ids = Array.from(selectedIds);
    try {
      await bulkComplete(ids);
      pushToast("success", `${ids.length} task${ids.length === 1 ? "" : "s"} marked complete.`);
      setSelectedIds(new Set());
      refetchStats();
    } catch (err) {
      pushToast("error", err instanceof ApiError ? err.message : "Failed to update tasks.");
    }
  }

  async function handleBulkDelete() {
    const ids = Array.from(selectedIds);
    try {
      await bulkDelete(ids);
      pushToast("success", `${ids.length} task${ids.length === 1 ? "" : "s"} deleted.`);
      setSelectedIds(new Set());
      refetchStats();
    } catch (err) {
      pushToast("error", err instanceof ApiError ? err.message : "Failed to delete tasks.");
    } finally {
      setBulkDeleteConfirm(false);
    }
  }

  useKeyboardShortcuts({
    onSearch: () => {
      setView("tasks");
      searchInputRef.current?.focus();
    },
    onNew: openAddForm,
  });

  const hasFilters = Boolean(debouncedSearch || status || priority || category || favoriteOnly);

  return (
    <div className="app-shell">
      <Sidebar
        view={view}
        onViewChange={setView}
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        taskCount={tasks.length}
      />

      <div className="app-content">
        <Header view={view} onAddTask={openAddForm} onOpenSidebar={() => setSidebarOpen(true)} />

        <main className="app-main">
          {view === "dashboard" ? (
            <Dashboard
              stats={stats}
              statsLoading={statsLoading}
              statsError={statsError}
              tasks={tasks}
              onViewTasks={() => setView("tasks")}
            />
          ) : (
            <>
              {error && <div className="banner-error">{error}</div>}

              <SearchFilterBar
                ref={searchInputRef}
                search={search}
                onSearchChange={setSearch}
                status={status}
                onStatusChange={setStatus}
                priority={priority}
                onPriorityChange={setPriority}
                category={category}
                onCategoryChange={setCategory}
                categories={categories}
                favoriteOnly={favoriteOnly}
                onFavoriteOnlyChange={setFavoriteOnly}
                sortBy={sortBy}
                onSortByChange={setSortBy}
                selectMode={selectMode}
                onToggleSelectMode={toggleSelectMode}
              />

              <BulkActionBar
                selectedCount={selectedIds.size}
                onComplete={handleBulkComplete}
                onDelete={() => setBulkDeleteConfirm(true)}
                onClear={() => setSelectedIds(new Set())}
              />

              <TaskList
                tasks={tasks}
                loading={loading}
                hasFilters={hasFilters}
                selectable={selectMode}
                selectedIds={selectedIds}
                onToggleSelect={toggleSelect}
                onEdit={openEditForm}
                onDelete={setDeletingTask}
                onDuplicate={handleDuplicate}
                onToggleFavorite={handleToggleFavorite}
                onArchive={handleArchive}
                onAddTask={openAddForm}
              />
            </>
          )}
        </main>
      </div>

      {formOpen && (
        <TaskForm initialTask={editingTask} categories={categories} onSubmit={handleFormSubmit} onClose={closeForm} />
      )}

      {deletingTask && (
        <ConfirmDialog
          title="Delete task"
          message={`Are you sure you want to delete "${deletingTask.title}"? This can't be undone.`}
          onConfirm={handleConfirmDelete}
          onCancel={() => setDeletingTask(null)}
        />
      )}

      {bulkDeleteConfirm && (
        <ConfirmDialog
          title="Delete selected tasks"
          message={`Are you sure you want to delete ${selectedIds.size} task${
            selectedIds.size === 1 ? "" : "s"
          }? This can't be undone.`}
          onConfirm={handleBulkDelete}
          onCancel={() => setBulkDeleteConfirm(false)}
        />
      )}

      <ToastStack toasts={toasts} />
    </div>
  );
}
