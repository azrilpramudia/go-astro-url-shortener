import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getLink } from "../lib/api";
import { loadCodes } from "../lib/storage";
import LinkManager from "./LinkManager";

vi.mock("../lib/api", () => ({
  ApiError: class extends Error {},
  createLink: vi.fn(),
  getLink: vi.fn(),
}));
vi.mock("../lib/storage", () => ({
  addCode: vi.fn(),
  loadCodes: vi.fn(),
  removeCode: vi.fn(),
}));

describe("LinkManager", () => {
  beforeEach(() => {
    vi.mocked(getLink).mockReset();
    vi.mocked(loadCodes).mockReset();
  });

  it("explains that the link list is local to this browser", async () => {
    vi.mocked(loadCodes).mockReturnValue([]);
    render(<LinkManager />);

    expect(
      await screen.findByText(
        "Your link list is stored in this browser only and won't sync across devices.",
      ),
    ).toBeTruthy();
  });

  it("shows a loading error instead of an empty-list message", async () => {
    vi.mocked(loadCodes).mockReturnValue(["missing-code"]);
    vi.mocked(getLink).mockRejectedValue(new Error("Not found"));
    render(<LinkManager />);

    expect(
      await screen.findByText("Unable to load your saved links. Please try refreshing."),
    ).toBeTruthy();
    expect(
      screen.queryByText("No links yet. Create your first one above."),
    ).toBeNull();
  });
});
