import { PUBLIC_API_URL } from "astro:env/client";
import type { ApiResponse, Link } from "../types/link";

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
      headers: { "Content-Type": "application/json", ...init?.headers },
    });
  } catch {
    throw new ApiError("Unable to connect to the server.", 0);
  }

  const body = (await res.json().catch(() => ({}))) as ApiResponse<T>;
  if (!res.ok || body.data === undefined) {
    throw new ApiError(body.errors ?? "Something went wrong.", res.status);
  }
  return body.data;
}

export function createLink(url: string, expiresAt?: string) {
  return request<Link>("/api/links", {
    method: "POST",
    body: JSON.stringify({ url, expires_at: expiresAt }),
  });
}

export function getLink(code: string) {
  return request<Link>(`/api/links/${encodeURIComponent(code)}`);
}
