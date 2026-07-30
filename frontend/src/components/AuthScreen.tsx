import { useState } from "react";
import type { FormEvent } from "react";
import { ApiError } from "../api/client";
import type { LoginInput, SignupInput } from "../types/task";

const NAME_MAX_LENGTH = 100;
const EMAIL_MAX_LENGTH = 150;
const PASSWORD_MIN_LENGTH = 8;
const PASSWORD_MAX_LENGTH = 72;
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

type Mode = "login" | "signup";

interface AuthScreenProps {
  onLogin: (input: LoginInput) => Promise<unknown>;
  onSignup: (input: SignupInput) => Promise<unknown>;
}

interface FormState {
  name: string;
  email: string;
  password: string;
}

interface FormErrors {
  name?: string;
  email?: string;
  password?: string;
}

/** Mirrors the backend's SignupRequest/LoginRequest.Validate rules so the
 * person gets instant feedback instead of a round trip. */
function validate(mode: Mode, state: FormState): FormErrors {
  const errors: FormErrors = {};

  if (mode === "signup") {
    const name = state.name.trim();
    if (!name) {
      errors.name = "Name is required.";
    } else if (name.length > NAME_MAX_LENGTH) {
      errors.name = `Name must be ${NAME_MAX_LENGTH} characters or fewer.`;
    }
  }

  const email = state.email.trim();
  if (!email) {
    errors.email = "Email is required.";
  } else if (email.length > EMAIL_MAX_LENGTH) {
    errors.email = `Email must be ${EMAIL_MAX_LENGTH} characters or fewer.`;
  } else if (!EMAIL_PATTERN.test(email)) {
    errors.email = "Enter a valid email address.";
  }

  if (!state.password) {
    errors.password = "Password is required.";
  } else if (mode === "signup" && state.password.length < PASSWORD_MIN_LENGTH) {
    errors.password = `Password must be at least ${PASSWORD_MIN_LENGTH} characters.`;
  } else if (state.password.length > PASSWORD_MAX_LENGTH) {
    errors.password = `Password must be ${PASSWORD_MAX_LENGTH} characters or fewer.`;
  }

  return errors;
}

export function AuthScreen({ onLogin, onSignup }: AuthScreenProps) {
  const [mode, setMode] = useState<Mode>("login");
  const [form, setForm] = useState<FormState>({ name: "", email: "", password: "" });
  const [errors, setErrors] = useState<FormErrors>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  function updateField<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  function switchMode(next: Mode) {
    setMode(next);
    setErrors({});
    setSubmitError(null);
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const validationErrors = validate(mode, form);
    setErrors(validationErrors);
    if (Object.keys(validationErrors).length > 0) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      const email = form.email.trim().toLowerCase();
      if (mode === "signup") {
        await onSignup({ name: form.name.trim(), email, password: form.password });
      } else {
        await onLogin({ email, password: form.password });
      }
    } catch (err) {
      setSubmitError(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");
      setSubmitting(false);
      return;
    }
    setSubmitting(false);
  }

  return (
    <div className="auth-shell">
      <div className="auth-card">
        <p className="app-header__eyebrow">Granet Technologies</p>
        <h1 className="auth-card__title">{mode === "login" ? "Log in" : "Create your account"}</h1>
        <p className="auth-card__subtitle">
          {mode === "login"
            ? "Welcome back. Log in to see your tasks."
            : "Track your tasks in one place — it only takes a minute to set up."}
        </p>

        <form onSubmit={handleSubmit} noValidate>
          {mode === "signup" && (
            <div className="form-field">
              <label htmlFor="auth-name">Name</label>
              <input
                id="auth-name"
                type="text"
                value={form.name}
                maxLength={NAME_MAX_LENGTH}
                className={errors.name ? "has-error" : ""}
                onChange={(e) => updateField("name", e.target.value)}
                autoFocus
              />
              {errors.name && <p className="form-field__error">{errors.name}</p>}
            </div>
          )}

          <div className="form-field">
            <label htmlFor="auth-email">Email</label>
            <input
              id="auth-email"
              type="email"
              value={form.email}
              maxLength={EMAIL_MAX_LENGTH}
              className={errors.email ? "has-error" : ""}
              onChange={(e) => updateField("email", e.target.value)}
              placeholder="name@example.com"
              autoFocus={mode === "login"}
              autoComplete="email"
            />
            {errors.email && <p className="form-field__error">{errors.email}</p>}
          </div>

          <div className="form-field">
            <label htmlFor="auth-password">Password</label>
            <input
              id="auth-password"
              type="password"
              value={form.password}
              maxLength={PASSWORD_MAX_LENGTH}
              className={errors.password ? "has-error" : ""}
              onChange={(e) => updateField("password", e.target.value)}
              placeholder={mode === "signup" ? "At least 8 characters" : "••••••••"}
              autoComplete={mode === "signup" ? "new-password" : "current-password"}
            />
            {errors.password && <p className="form-field__error">{errors.password}</p>}
          </div>

          {submitError && <p className="form-field__error">{submitError}</p>}

          <button type="submit" className="btn btn-accent auth-card__submit" disabled={submitting}>
            {submitting
              ? mode === "login"
                ? "Logging in…"
                : "Creating account…"
              : mode === "login"
                ? "Log in"
                : "Create account"}
          </button>
        </form>

        <p className="auth-card__toggle">
          {mode === "login" ? (
            <>
              New here?{" "}
              <button type="button" className="auth-card__link" onClick={() => switchMode("signup")}>
                Create an account
              </button>
            </>
          ) : (
            <>
              Already have an account?{" "}
              <button type="button" className="auth-card__link" onClick={() => switchMode("login")}>
                Log in
              </button>
            </>
          )}
        </p>
      </div>
    </div>
  );
}
