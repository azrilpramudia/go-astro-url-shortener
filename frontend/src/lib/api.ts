import { PUBLIC_API_URL } from "astro:env/client";
import type {
  ApiResponse,
  Link,
  LinkAnalytics,
  LinkListFilters,
  LinkPage,
  User,
} from "../types/link";

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
    throw new ApiError("Unable to connect to the server.", 0);
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

export const createLink = (
  url: string,
  expiresAt?: string,
  customCode?: string,
  tags: string[] = [],
) =>
  request<Link>("/api/links", {
    method: "POST",
    body: JSON.stringify({
      url,
      expires_at: expiresAt,
      custom_code: customCode,
      tags,
    }),
  });

export const listMyLinks = (
  page = 1,
  size = 20,
  filters: LinkListFilters = {},
) => {
  const params = new URLSearchParams({ page: String(page), size: String(size) });
  if (filters.q?.trim()) params.set("q", filters.q.trim());
  if (filters.status) params.set("status", filters.status);
  if (filters.tag?.trim()) params.set("tag", filters.tag.trim());
  return request<LinkPage>(`/api/me/links?${params.toString()}`);
};

export const getLinkAnalytics = (code: string, days = 30) =>
  request<LinkAnalytics>(
    `/api/me/links/${encodeURIComponent(code)}/analytics?days=${days}`,
  );

export const updateLink = (
  code: string,
  updates: { url: string; expires_at?: string; tags: string[] },
) =>
  request<Link>(`/api/links/${encodeURIComponent(code)}`, {
    method: "PUT",
    body: JSON.stringify(updates),
  });

export const deleteLink = (code: string) =>
  request<void>(`/api/links/${encodeURIComponent(code)}`, { method: "DELETE" });
