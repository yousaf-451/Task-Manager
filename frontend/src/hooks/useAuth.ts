import { useCallback, useEffect, useState } from "react";
import { authApi } from "../api/authApi";
import { ApiError } from "../api/client";
import type { LoginInput, SignupInput, User } from "../types/task";

interface UseAuthResult {
  user: User | null;
  /** True only while the initial "is there already a session?" check runs. */
  checkingSession: boolean;
  signup: (input: SignupInput) => Promise<User>;
  login: (input: LoginInput) => Promise<User>;
  logout: () => Promise<void>;
  deleteAccount: () => Promise<void>;
}

/**
 * Tracks the signed-in user (or the lack of one) and exposes the auth
 * actions. On mount it calls GET /api/auth/me once to find out whether the
 * browser already has a valid session cookie, so a page refresh doesn't
 * force a fresh login.
 */
export function useAuth(): UseAuthResult {
  const [user, setUser] = useState<User | null>(null);
  const [checkingSession, setCheckingSession] = useState(true);

  useEffect(() => {
    let cancelled = false;
    authApi
      .me()
      .then((u) => {
        if (!cancelled) setUser(u);
      })
      .catch(() => {
        // No session (401) or a network hiccup - either way, treat as
        // "not logged in" and let the person sign up / log in.
        if (!cancelled) setUser(null);
      })
      .finally(() => {
        if (!cancelled) setCheckingSession(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const signup = useCallback(async (input: SignupInput) => {
    const u = await authApi.signup(input);
    setUser(u);
    return u;
  }, []);

  const login = useCallback(async (input: LoginInput) => {
    const u = await authApi.login(input);
    setUser(u);
    return u;
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch (err) {
      // Even if the server call fails (e.g. the session was already
      // invalid), clear the local user so the UI reflects "logged out" -
      // there's nothing useful to retry here.
      if (!(err instanceof ApiError)) throw err;
    } finally {
      setUser(null);
    }
  }, []);

  const deleteAccount = useCallback(async () => {
    await authApi.deleteAccount();
    setUser(null);
  }, []);

  return { user, checkingSession, signup, login, logout, deleteAccount };
}
