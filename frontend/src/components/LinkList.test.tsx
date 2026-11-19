import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { copyToClipboard } from "../lib/clipboard";
import { getLinkAnalytics, updateLink } from "../lib/api";
import type { Link } from "../types/link";
import LinkList from "./LinkList";

vi.mock("../lib/clipboard", () => ({ copyToClipboard: vi.fn() }));
vi.mock("../lib/api", () => ({
  getLinkAnalytics: vi.fn(),
  updateLink: vi.fn(),
}));

const link: Link = {
  id: 1,
  code: "abc1234",
  short_url: "http://localhost:3000/abc1234",
  original_url: "https://example.com/a-long-url",
  clicks: 3,
  expires_at: "2026-12-01T10:00:00Z",
  created_at: "2026-11-18T00:00:00Z",
  tags: [],
};

describe("LinkList", () => {
  beforeEach(() => {
    vi.mocked(copyToClipboard).mockReset();
    vi.mocked(getLinkAnalytics).mockReset();
    vi.mocked(updateLink).mockReset();
  });

  it("shows the link expiration date", () => {
    render(<LinkList links={[link]} onRemove={vi.fn()} />);

    expect(screen.getByText(/^Expires /).textContent).not.toContain("Invalid Date");
    expect(screen.queryByRole("button", { name: "Analytics" })).toBeNull();
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

  it("loads the selected link's analytics and supports changing the date range", async () => {
    vi.mocked(getLinkAnalytics).mockResolvedValue({
      days: 30,
      since: "2026-10-19T00:00:00Z",
      total_clicks: 25,
      unique_visitors: 18,
      location_available: false,
      clicks_by_day: [{ date: "2026-11-18", clicks: 4 }],
      referrers: [{ name: "Direct", count: 4 }],
      countries: [],
      devices: [{ name: "desktop", count: 3 }],
    });
    render(<LinkList links={[link]} analyticsEnabled />);

    fireEvent.click(screen.getByRole("button", { name: "Analytics" }));

    expect(await screen.findByText("Link analytics")).toBeTruthy();
    expect(screen.getByText("25")).toBeTruthy();
    expect(screen.getByText("18")).toBeTruthy();
    expect(screen.getByRole("img", { name: "Daily click trend" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Export CSV" })).toBeTruthy();
    expect(
      screen.getByText(
        "Country analytics are unavailable until a local GeoIP database is configured.",
      ),
    ).toBeTruthy();
    expect(getLinkAnalytics).toHaveBeenCalledWith("abc1234", 30);

    fireEvent.click(screen.getByRole("button", { name: "7 days" }));
    expect(await screen.findByRole("button", { name: "7 days" })).toBeTruthy();
    expect(getLinkAnalytics).toHaveBeenLastCalledWith("abc1234", 7);
  });

  it("shows an SVG QR code for the short link", () => {
    const { container } = render(<LinkList links={[link]} />);
    fireEvent.click(screen.getByRole("button", { name: "Show QR code abc1234" }));

    expect(screen.getByRole("region", { name: "QR code for abc1234" })).toBeTruthy();
    expect(container.querySelector("svg")).toBeTruthy();
    expect(screen.getByRole("button", { name: "Download QR as SVG" })).toBeTruthy();
  });

  it("updates a link and reports the result to its owner list", async () => {
    const updated = { ...link, original_url: "https://updated.example", tags: ["launch"] };
    const onUpdate = vi.fn();
    vi.mocked(updateLink).mockResolvedValue(updated);
    render(<LinkList links={[link]} onUpdate={onUpdate} />);

    fireEvent.click(screen.getByRole("button", { name: "Edit abc1234" }));
    fireEvent.change(screen.getByLabelText("Destination URL"), {
      target: { value: updated.original_url },
    });
    fireEvent.change(screen.getByLabelText(/Tags/), {
      target: { value: "launch" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }));

    await waitFor(() => expect(onUpdate).toHaveBeenCalledWith(updated));
    expect(screen.queryByRole("form", { name: "Edit link abc1234" })).toBeNull();
  });
});
