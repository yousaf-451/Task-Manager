import { apiClient } from "./client";
import type { LoginInput, SignupInput, User } from "../types/task";

export const authApi = {
  signup: (input: SignupInput) => apiClient.post<User>("/auth/signup", input),
  login: (input: LoginInput) => apiClient.post<User>("/auth/login", input),
  logout: () => apiClient.post<{ loggedOut: boolean }>("/auth/logout", {}),
  me: () => apiClient.get<User>("/auth/me"),
  deleteAccount: () => apiClient.delete<{ deleted: boolean }>("/auth/me"),
};
