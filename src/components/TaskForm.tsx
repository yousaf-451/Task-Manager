import { useState } from "react";
import type { FormEvent } from "react";
import { Modal } from "./Modal";
import type { CreateTaskInput, Task, TaskPriority, TaskStatus } from "../types/task";
import { PRIORITY_LABELS, STATUS_LABELS, TASK_COLORS, TASK_PRIORITIES, TASK_STATUSES } from "../types/task";
import { TASK_TITLE_MAX_LENGTH, TASK_DESCRIPTION_MAX_LENGTH } from "../constants";

interface TaskFormProps {
  /** When editing, the existing task; when creating, undefined. */
  initialTask?: Task;
  categories: string[];
  onSubmit: (input: CreateTaskInput) => Promise<void>;
  onClose: () => void;
}

interface FormState {
  title: string;
  description: string;
  dueDate: string;
  status: TaskStatus;
  priority: TaskPriority;
  category: string;
  color: string;
}

interface FormErrors {
  title?: string;
  description?: string;
  dueDate?: string;
}

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

function toFormState(task?: Task): FormState {
  if (!task) {
    return {
      title: "",
      description: "",
      dueDate: todayISO(),
      status: "pending",
      priority: "medium",
      category: "",
      color: TASK_COLORS[0],
    };
  }
  return {
    title: task.title,
    description: task.description,
    dueDate: task.dueDate,
    status: task.status,
    priority: task.priority,
    category: task.category,
    color: task.color,
  };
}

/** Validates the form client-side, mirroring the backend's rules so
 * the user gets instant feedback instead of a round trip. */
function validate(state: FormState): FormErrors {
  const errors: FormErrors = {};

  const title = state.title.trim();
  if (!title) {
    errors.title = "Title is required.";
  } else if (title.length > TASK_TITLE_MAX_LENGTH) {
    errors.title = `Title must be ${TASK_TITLE_MAX_LENGTH} characters or fewer.`;
  }

  if (state.description.trim().length > TASK_DESCRIPTION_MAX_LENGTH) {
    errors.description = `Description must be ${TASK_DESCRIPTION_MAX_LENGTH} characters or fewer.`;
  }

  if (!state.dueDate) {
    errors.dueDate = "Due date is required.";
  }

  return errors;
}

export function TaskForm({ initialTask, categories, onSubmit, onClose }: TaskFormProps) {
  const [form, setForm] = useState<FormState>(() => toFormState(initialTask));
  const [errors, setErrors] = useState<FormErrors>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const isEditing = Boolean(initialTask);

  function updateField<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const validationErrors = validate(form);
    setErrors(validationErrors);
    if (Object.keys(validationErrors).length > 0) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      await onSubmit({
        title: form.title.trim(),
        description: form.description.trim(),
        dueDate: form.dueDate,
        status: form.status,
        priority: form.priority,
        category: form.category.trim(),
        color: form.color,
      });
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : "Something went wrong. Please try again.");
      setSubmitting(false);
      return;
    }
    setSubmitting(false);
  }

  return (
    <Modal title={isEditing ? "Edit task" : "Add task"} onClose={onClose}>
      <form onSubmit={handleSubmit} noValidate>
        <div className="form-field">
          <label htmlFor="title">Title</label>
          <input
            id="title"
            type="text"
            value={form.title}
            maxLength={TASK_TITLE_MAX_LENGTH}
            className={errors.title ? "has-error" : ""}
            onChange={(e) => updateField("title", e.target.value)}
            autoFocus
          />
          {errors.title && <p className="form-field__error">{errors.title}</p>}
        </div>

        <div className="form-field">
          <label htmlFor="description">Description</label>
          <textarea
            id="description"
            value={form.description}
            maxLength={TASK_DESCRIPTION_MAX_LENGTH}
            className={errors.description ? "has-error" : ""}
            onChange={(e) => updateField("description", e.target.value)}
            placeholder="Optional details about this task"
          />
          {errors.description && <p className="form-field__error">{errors.description}</p>}
        </div>

        <div className="form-row">
          <div className="form-field">
            <label htmlFor="dueDate">Due date</label>
            <input
              id="dueDate"
              type="date"
              value={form.dueDate}
              className={errors.dueDate ? "has-error" : ""}
              onChange={(e) => updateField("dueDate", e.target.value)}
            />
            {errors.dueDate && <p className="form-field__error">{errors.dueDate}</p>}
          </div>

          <div className="form-field">
            <label htmlFor="status">Status</label>
            <select
              id="status"
              value={form.status}
              onChange={(e) => updateField("status", e.target.value as TaskStatus)}
            >
              {TASK_STATUSES.map((s) => (
                <option key={s} value={s}>
                  {STATUS_LABELS[s]}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="form-row">
          <div className="form-field">
            <label htmlFor="priority">Priority</label>
            <select
              id="priority"
              value={form.priority}
              onChange={(e) => updateField("priority", e.target.value as TaskPriority)}
            >
              {TASK_PRIORITIES.map((p) => (
                <option key={p} value={p}>
                  {PRIORITY_LABELS[p]}
                </option>
              ))}
            </select>
          </div>

          <div className="form-field">
            <label htmlFor="category">Category</label>
            <input
              id="category"
              type="text"
              list="category-suggestions"
              value={form.category}
              placeholder="e.g. Backend"
              onChange={(e) => updateField("category", e.target.value)}
            />
            <datalist id="category-suggestions">
              {categories.map((c) => (
                <option key={c} value={c} />
              ))}
            </datalist>
          </div>
        </div>

        <div className="form-field">
          <label>Color</label>
          <div className="color-picker" role="radiogroup" aria-label="Task color">
            {TASK_COLORS.map((c) => (
              <button
                key={c}
                type="button"
                role="radio"
                aria-checked={form.color === c}
                aria-label={`Color ${c}`}
                className={`color-swatch ${form.color === c ? "color-swatch--selected" : ""}`}
                style={{ background: c }}
                onClick={() => updateField("color", c)}
              />
            ))}
          </div>
        </div>

        {submitError && <p className="form-field__error">{submitError}</p>}

        <div className="modal__actions">
          <button type="button" className="btn btn-ghost" onClick={onClose} disabled={submitting}>
            Cancel
          </button>
          <button type="submit" className="btn btn-accent" disabled={submitting}>
            {submitting ? "Saving…" : isEditing ? "Save changes" : "Add task"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
