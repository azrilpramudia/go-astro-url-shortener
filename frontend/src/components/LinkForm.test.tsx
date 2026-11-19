import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError, createLink } from "../lib/api";
import { copyToClipboard } from "../lib/clipboard";
import type { Link } from "../types/link";
import LinkForm from "./LinkForm";

vi.mock("../lib/api", () => ({
  ApiError: class extends Error {
    status: number;

    constructor(message: string, status: number) {
      super(message);
      this.status = status;
    }
  },
  createLink: vi.fn(),
}));
vi.mock("../lib/clipboard", () => ({ copyToClipboard: vi.fn() }));

const link: Link = {
  id: 1,
  code: "abc1234",
  short_url: "http://localhost:3000/abc1234",
  original_url: "https://example.com/a-long-url",
  clicks: 0,
  created_at: "2026-11-18T00:00:00Z",
  tags: [],
};

describe("LinkForm", () => {
  beforeEach(() => {
    vi.mocked(createLink).mockReset();
    vi.mocked(copyToClipboard).mockReset();
  });

  it("shows the created short link and copy feedback", async () => {
    const onCreated = vi.fn();
    vi.mocked(createLink).mockResolvedValue(link);
    vi.mocked(copyToClipboard).mockResolvedValue();

    render(<LinkForm onCreated={onCreated} />);
    fireEvent.change(screen.getByLabelText("Long URL"), {
      target: { value: link.original_url },
    });
    fireEvent.click(screen.getByRole("button", { name: "Shorten" }));

    expect(await screen.findByText("Short link created successfully.")).toBeTruthy();
    expect(screen.getByRole("link", { name: link.short_url })).toBeTruthy();
    expect(onCreated).toHaveBeenCalledWith(link);

    fireEvent.click(screen.getByRole("button", { name: "Copy short link" }));
    expect(await screen.findByText("Copied to clipboard.")).toBeTruthy();
    expect(copyToClipboard).toHaveBeenCalledWith(link.short_url);
  });

  it("shows API errors when link creation fails", async () => {
    vi.mocked(createLink).mockRejectedValue(new ApiError("Invalid URL.", 400));

    render(<LinkForm onCreated={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Long URL"), {
      target: { value: "https://example.com" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Shorten" }));

    expect((await screen.findByRole("alert")).textContent).toBe("Invalid URL.");
  });

  it("submits a custom alias and tags for an authenticated user", async () => {
    vi.mocked(createLink).mockResolvedValue({ ...link, tags: ["campaign"] });
    render(<LinkForm onCreated={vi.fn()} authenticated />);
    fireEvent.change(screen.getByLabelText("Long URL"), {
      target: { value: link.original_url },
    });
    fireEvent.change(screen.getByLabelText(/Custom alias/), {
      target: { value: "my-link" },
    });
    fireEvent.change(screen.getByLabelText(/Tags/), {
      target: { value: "campaign" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Shorten" }));

    expect(await screen.findByText("Short link created successfully.")).toBeTruthy();
    expect(createLink).toHaveBeenCalledWith(
      link.original_url,
      undefined,
      "my-link",
      ["campaign"],
    );
  });
});
