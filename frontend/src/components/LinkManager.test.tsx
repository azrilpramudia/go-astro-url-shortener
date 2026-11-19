import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError, getMe, listMyLinks } from "../lib/api";
import type { Link, LinkPage, User } from "../types/link";
import LinkManager from "./LinkManager";

vi.mock("../lib/api", () => ({
  ApiError: class extends Error {
    status: number;

    constructor(message: string, status: number) {
      super(message);
      this.status = status;
    }
  },
  deleteLink: vi.fn(),
  getMe: vi.fn(),
  listMyLinks: vi.fn(),
  logout: vi.fn(),
}));

const user: User = {
  id: 1,
  name: "Alex",
  email: "alex@example.com",
};

const link: Link = {
  id: 1,
  code: "abc1234",
  short_url: "http://localhost:3000/abc1234",
  original_url: "https://example.com/a-long-url",
  clicks: 3,
  created_at: "2026-11-18T00:00:00Z",
};

const linkPage: LinkPage = {
  items: [link],
  total: 1,
  page: 1,
  size: 20,
};

describe("LinkManager", () => {
  beforeEach(() => {
    vi.mocked(getMe).mockReset();
    vi.mocked(listMyLinks).mockReset();
  });

  it("shows sign-in actions when no user is authenticated", async () => {
    vi.mocked(getMe).mockRejectedValue(new ApiError("Unauthorized", 401));
    render(<LinkManager />);

    expect(await screen.findByRole("link", { name: "Sign in" })).toBeTruthy();
    expect(screen.getByRole("link", { name: "Sign up" })).toBeTruthy();
    expect(screen.getByRole("heading", { name: "Recent links" })).toBeTruthy();
  });

  it("loads and displays links for the authenticated user", async () => {
    vi.mocked(getMe).mockResolvedValue(user);
    vi.mocked(listMyLinks).mockResolvedValue(linkPage);
    render(<LinkManager />);

    expect(
      await screen.findByRole("heading", { name: "Your links" }),
    ).toBeTruthy();
    expect(screen.getByRole("link", { name: link.short_url })).toBeTruthy();
  });
});
