import { PUBLIC_API_URL } from "astro:env/client";
import type { ApiResponse, Link, LinkPage, User } from "../types/link";

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${PUBLIC_API_URL}${path}`, {
      ...init,
      credentials: "include",
      headers: { "Content-Type": "application/json", ...init?.headers },
    });
  } catch {
    throw new ApiError("Tidak bisa terhubung ke server", 0);
  }

  if (res.status === 204) return undefined as T;

  const body = (await res.json().catch(() => ({}))) as ApiResponse<T>;
  if (!res.ok || body.data === undefined) {
    throw new ApiError(body.errors ?? "Terjadi kesalahan", res.status);
  }
  return body.data;
}

export const register = (name: string, email: string, password: string) =>
  request<User>("/api/auth/register", {
    method: "POST",
    body: JSON.stringify({ name, email, password }),
  });

export const login = (email: string, password: string) =>
  request<User>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });

export const logout = () =>
  request<void>("/api/auth/logout", { method: "POST" });

export const getMe = () => request<User>("/api/auth/me");

export const createLink = (url: string, expiresAt?: string) =>
  request<Link>("/api/links", {
    method: "POST",
    body: JSON.stringify({ url, expires_at: expiresAt }),
  });

export const listMyLinks = (page = 1, size = 50) =>
  request<LinkPage>(`/api/me/links?page=${page}&size=${size}`);

export const deleteLink = (code: string) =>
  request<void>(`/api/links/${encodeURIComponent(code)}`, { method: "DELETE" });
