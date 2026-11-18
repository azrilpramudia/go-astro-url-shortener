import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { copyToClipboard } from "../lib/clipboard";
import type { Link } from "../types/link";
import LinkList from "./LinkList";

vi.mock("../lib/clipboard", () => ({ copyToClipboard: vi.fn() }));

const link: Link = {
  id: 1,
  code: "abc1234",
  short_url: "http://localhost:3000/abc1234",
  original_url: "https://example.com/a-long-url",
  clicks: 3,
  expires_at: "2026-12-01T10:00:00Z",
  created_at: "2026-11-18T00:00:00Z",
};

describe("LinkList", () => {
  beforeEach(() => {
    vi.mocked(copyToClipboard).mockReset();
  });

  it("shows the link expiration date", () => {
    render(<LinkList links={[link]} onRemove={vi.fn()} />);

    expect(screen.getByText(/^Expires /).textContent).not.toContain("Invalid Date");
  });

  it("shows an error when the clipboard write fails", async () => {
    vi.mocked(copyToClipboard).mockRejectedValue(new Error("Permission denied"));
    render(<LinkList links={[link]} onRemove={vi.fn()} />);

    fireEvent.click(
      screen.getByRole("button", { name: "Copy short link abc1234" }),
    );

    expect(
      await screen.findByText("Could not copy the link. Please copy it manually."),
    ).toBeTruthy();
  });
});
