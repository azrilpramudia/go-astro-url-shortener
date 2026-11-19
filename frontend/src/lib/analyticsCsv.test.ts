import { describe, expect, it } from "vitest";
import { buildAnalyticsCsv } from "./analyticsCsv";

describe("buildAnalyticsCsv", () => {
  it("exports the selected analytics and neutralizes spreadsheet formulas", () => {
    const csv = buildAnalyticsCsv({
      days: 7,
      since: "2026-11-13T00:00:00Z",
      total_clicks: 3,
      unique_visitors: 2,
      location_available: false,
      clicks_by_day: [{ date: "2026-11-19", clicks: 3 }],
      referrers: [{ name: '=HYPERLINK("https://bad.example")', count: 3 }],
      countries: [],
      devices: [{ name: "mobile", count: 3 }],
    });

    expect(csv).toContain('"summary","tracked clicks","3"');
    expect(csv).toContain('"daily clicks","2026-11-19","3"');
    expect(csv).toContain(
      `"'=HYPERLINK(""https://bad.example"")","3"`,
    );
    expect(csv.startsWith("\uFEFF")).toBe(true);
  });
});
