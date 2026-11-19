import type { LinkAnalytics } from "../types/link";

export function buildAnalyticsCsv(analytics: LinkAnalytics) {
  const rows: Array<Array<string | number>> = [
    ["section", "name", "count"],
    ["summary", "tracked clicks", analytics.total_clicks],
    ["summary", "unique visitors", analytics.unique_visitors],
    ["summary", "range days", analytics.days],
    ["summary", "since", analytics.since],
    ...analytics.clicks_by_day.map((day) => ["daily clicks", day.date, day.clicks]),
    ...analytics.referrers.map((item) => ["referrer", item.name, item.count]),
    ...analytics.countries.map((item) => ["country", item.name, item.count]),
    ...analytics.devices.map((item) => ["device", item.name, item.count]),
  ];
  return `\uFEFF${rows
    .map((row) => row.map(csvCell).join(","))
    .join("\r\n")}`;
}

function csvCell(value: string | number) {
  let text = String(value);
  if (/^\s*[=+\-@]/.test(text)) text = `'${text}`;
  return `"${text.replaceAll('"', '""')}"`;
}
