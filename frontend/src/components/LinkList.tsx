import { useEffect, useRef, useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { copyToClipboard } from "../lib/clipboard";
import { getLinkAnalytics } from "../lib/api";
import { buildAnalyticsCsv } from "../lib/analyticsCsv";
import type { Link, LinkAnalytics } from "../types/link";
import EditLinkForm from "./EditLinkForm";

interface Props {
  links: Link[];
  onRemove?: (code: string) => void;
  analyticsEnabled?: boolean;
  onUpdate?: (link: Link) => void;
}

export default function LinkList({
  links,
  onRemove,
  analyticsEnabled = false,
  onUpdate,
}: Props) {
  const [analyticsCode, setAnalyticsCode] = useState<string | null>(null);
  const [qrCode, setQrCode] = useState<string | null>(null);
  const [editingCode, setEditingCode] = useState<string | null>(null);
  const [copyFeedback, setCopyFeedback] = useState<{
    code: string;
    status: "copied" | "error";
  } | null>(null);

  async function copy(link: Link) {
    setCopyFeedback(null);
    try {
      await copyToClipboard(link.short_url);
      setCopyFeedback({ code: link.code, status: "copied" });
    } catch {
      setCopyFeedback({ code: link.code, status: "error" });
    }
  }

  if (links.length === 0) {
    return (
      <div className="px-5 py-10 text-center sm:py-12">
        <p className="font-medium text-slate-800">No links yet</p>
        <p className="mt-1 text-sm text-slate-500">
          Your short links will appear here.
        </p>
      </div>
    );
  }

  return (
    <ul className="divide-y divide-slate-100">
      {links.map((link) => {
        const expiration = link.expires_at
          ? formatExpiration(link.expires_at)
          : null;
        const tags = link.tags ?? [];
        const feedback =
          copyFeedback?.code === link.code ? copyFeedback.status : null;

        return (
          <li
            key={link.code}
            className="flex flex-col gap-3 p-4 transition-colors hover:bg-slate-50/70 sm:flex-row sm:items-center sm:p-5"
          >
            <div className="min-w-0 flex-1">
              <a
                href={link.short_url}
                target="_blank"
                rel="noopener noreferrer"
                className="block break-all font-semibold text-indigo-700 decoration-indigo-300 underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:break-normal sm:truncate"
              >
                {link.short_url}
              </a>
              <p className="break-all text-sm text-slate-500 sm:truncate sm:break-normal">
                {link.original_url}
              </p>
              {expiration && (
                <p className="text-sm text-slate-600">Expires {expiration}</p>
              )}
              {tags.length > 0 && (
                <ul aria-label={`Tags for ${link.code}`} className="mt-1 flex flex-wrap gap-1">
                  {tags.map((tag) => (
                    <li
                      key={tag}
                      className="rounded-full bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-700"
                    >
                      {tag}
                    </li>
                  ))}
                </ul>
              )}
            </div>
            <div className="flex flex-col gap-2 sm:items-end">
              <div className="flex flex-wrap items-center gap-2 sm:flex-nowrap">
                <span className="mr-auto whitespace-nowrap text-sm text-slate-600 sm:mr-0">
                  {link.clicks} clicks
                </span>
                {analyticsEnabled && (
                  <button
                    type="button"
                    aria-expanded={analyticsCode === link.code}
                    aria-controls={
                      analyticsCode === link.code
                        ? `analytics-${link.code}`
                        : undefined
                    }
                    onClick={() =>
                      setAnalyticsCode((current) =>
                        current === link.code ? null : link.code,
                      )
                    }
                    className="inline-flex min-h-10 items-center justify-center rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                  >
                    {analyticsCode === link.code
                      ? "Hide analytics"
                      : "Analytics"}
                  </button>
                )}
                {onUpdate && (
                  <button
                    type="button"
                    aria-label={`Edit ${link.code}`}
                    onClick={() =>
                      setEditingCode((current) =>
                        current === link.code ? null : link.code,
                      )
                    }
                    className="min-h-10 rounded-lg px-3 text-sm font-medium text-slate-600 transition hover:bg-slate-100"
                  >
                    Edit
                  </button>
                )}
                <button
                  type="button"
                  aria-expanded={qrCode === link.code}
                  aria-label={`Show QR code ${link.code}`}
                  onClick={() =>
                    setQrCode((current) =>
                      current === link.code ? null : link.code,
                    )
                  }
                  className="min-h-10 rounded-lg px-3 text-sm font-medium text-slate-600 transition hover:bg-slate-100"
                >
                  QR
                </button>
                <button
                  type="button"
                  onClick={() => copy(link)}
                  aria-label={`Copy short link ${link.code}`}
                  className="inline-flex min-h-10 flex-1 items-center justify-center rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:flex-none"
                >
                  {feedback === "copied" ? "Copied" : "Copy"}
                </button>
                {onRemove && (
                  <button
                    onClick={() => onRemove(link.code)}
                    aria-label={`Remove ${link.code}`}
                    className="min-h-10 rounded-lg px-3 text-sm font-medium text-slate-600 transition hover:bg-red-50 hover:text-red-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                  >
                    Remove
                  </button>
                )}
              </div>
              {analyticsEnabled && analyticsCode === link.code && (
                <div id={`analytics-${link.code}`} className="w-full sm:w-[min(100%,36rem)]">
                  <LinkAnalyticsPanel code={link.code} />
                </div>
              )}
              {editingCode === link.code && onUpdate && (
                <div className="w-full sm:w-[min(100%,36rem)]">
                  <EditLinkForm
                    link={link}
                    onCancel={() => setEditingCode(null)}
                    onUpdated={(updated) => {
                      onUpdate(updated);
                      setEditingCode(null);
                    }}
                  />
                </div>
              )}
              {qrCode === link.code && (
                <div className="w-full sm:w-[min(100%,36rem)]">
                  <QRCodePanel link={link} />
                </div>
              )}
              {feedback === "copied" && (
                <p role="status" className="text-sm text-emerald-700">
                  Copied to clipboard.
                </p>
              )}
              {feedback === "error" && (
                <p role="alert" className="text-sm text-red-700">
                  Could not copy the link. Please copy it manually.
                </p>
              )}
            </div>
          </li>
        );
      })}
    </ul>
  );
}

function LinkAnalyticsPanel({ code }: { code: string }) {
  const [days, setDays] = useState(30);
  const [analytics, setAnalytics] = useState<LinkAnalytics | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [exportError, setExportError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);
    setAnalytics(null);
    getLinkAnalytics(code, days)
      .then((result) => {
        if (active) setAnalytics(result);
      })
      .catch((cause: unknown) => {
        if (active) {
          setError(
            cause instanceof Error
              ? cause.message
              : "Could not load analytics. Please try again.",
          );
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [code, days]);

  return (
    <section
      aria-label={`Analytics for ${code}`}
      className="mt-2 rounded-xl border border-slate-200 bg-white p-4 text-left shadow-sm"
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h3 className="font-semibold text-slate-900">Link analytics</h3>
        <div className="flex flex-wrap items-center gap-2">
          {analytics && (
            <button
              type="button"
              onClick={() => {
                try {
                  downloadAnalyticsCsv(code, analytics.days, buildAnalyticsCsv(analytics));
                  setExportError(null);
                } catch {
                  setExportError("Could not export analytics. Please try again.");
                }
              }}
              className="rounded-md border border-slate-200 px-2.5 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50"
            >
              Export CSV
            </button>
          )}
          <div aria-label="Analytics date range" className="flex gap-1">
            {[7, 30, 90].map((range) => (
              <button
                key={range}
                type="button"
                aria-pressed={days === range}
                onClick={() => setDays(range)}
                className={`rounded-md px-2.5 py-1.5 text-xs font-medium ${
                  days === range
                    ? "bg-indigo-100 text-indigo-800"
                    : "text-slate-600 hover:bg-slate-100"
                }`}
              >
                {range} days
              </button>
            ))}
          </div>
        </div>
      </div>

      {loading && (
        <p role="status" className="mt-4 text-sm text-slate-500">
          Loading analytics…
        </p>
      )}
      {error && (
        <p role="alert" className="mt-4 text-sm text-red-700">
          {error}
        </p>
      )}
      {exportError && (
        <p role="alert" className="mt-3 text-sm text-red-700">
          {exportError}
        </p>
      )}
      {!loading && analytics && (
        <>
          <div className="mt-4 grid grid-cols-2 gap-3">
            <Metric label="Tracked clicks" value={analytics.total_clicks} />
            <Metric label="Unique visitors" value={analytics.unique_visitors} />
          </div>
          <ClickTrend clicks={analytics.clicks_by_day} />
          <div className="mt-4 grid gap-4 sm:grid-cols-3">
            <Breakdown title="Top referrers" items={analytics.referrers} />
            <Breakdown title="Countries" items={analytics.countries} />
            <Breakdown title="Devices" items={analytics.devices} />
          </div>
          {!analytics.location_available && (
            <p className="mt-4 text-xs text-slate-500">
              Country analytics are unavailable until a local GeoIP database is configured.
            </p>
          )}
          <p className="mt-3 text-xs text-slate-400">
            Unique visitors are estimated using a first-party browser cookie. Analytics events are
            removed according to the configured retention policy.
          </p>
        </>
      )}
    </section>
  );
}

function downloadAnalyticsCsv(code: string, days: number, csv: string) {
  const objectUrl = URL.createObjectURL(
    new Blob([csv], { type: "text/csv;charset=utf-8" }),
  );
  const anchor = document.createElement("a");
  anchor.href = objectUrl;
  anchor.download = `analytics-${code}-${days}d.csv`;
  anchor.click();
  window.setTimeout(() => URL.revokeObjectURL(objectUrl), 0);
}

function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg bg-slate-50 px-3 py-2">
      <p className="text-xs font-medium text-slate-500">{label}</p>
      <p className="mt-1 text-xl font-semibold tabular-nums text-slate-900">
        {value.toLocaleString()}
      </p>
    </div>
  );
}

function QRCodePanel({ link }: { link: Link }) {
  const svgRef = useRef<SVGSVGElement>(null);

  function download() {
    if (!svgRef.current) return;
    const markup = new XMLSerializer().serializeToString(svgRef.current);
    const anchor = document.createElement("a");
    anchor.href = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(markup)}`;
    anchor.download = `${link.code}-qr.svg`;
    anchor.click();
  }

  return (
    <section
      aria-label={`QR code for ${link.code}`}
      className="mt-2 flex flex-col items-center gap-3 rounded-xl border border-slate-200 bg-white p-4"
    >
      <QRCodeSVG ref={svgRef} value={link.short_url} size={192} marginSize={4} />
      <p className="max-w-full break-all text-center text-xs text-slate-500">
        {link.short_url}
      </p>
      <button
        type="button"
        onClick={download}
        className="min-h-10 rounded-lg border border-slate-200 px-3 text-sm font-medium text-slate-700 hover:bg-slate-50"
      >
        Download QR as SVG
      </button>
    </section>
  );
}

function ClickTrend({ clicks }: { clicks: LinkAnalytics["clicks_by_day"] }) {
  const width = 600;
  const height = 150;
  const padding = 12;
  const maxClicks = Math.max(...clicks.map((day) => day.clicks), 1);
  const points = clicks
    .map((day, index) => {
      const x =
        clicks.length <= 1
          ? width / 2
          : (index / (clicks.length - 1)) * (width - 2 * padding) + padding;
      const y =
        height -
        padding -
        (day.clicks / maxClicks) * (height - 2 * padding);
      return `${x},${y}`;
    })
    .join(" ");

  return (
    <div className="mt-4">
      <h4 className="text-sm font-medium text-slate-700">Clicks over time</h4>
      {clicks.length === 0 || clicks.every((day) => day.clicks === 0) ? (
        <p className="mt-2 text-sm text-slate-500">No clicks in this period.</p>
      ) : (
        <svg
          viewBox={`0 0 ${width} ${height}`}
          role="img"
          aria-label="Daily click trend"
          className="mt-2 h-32 w-full overflow-visible"
          preserveAspectRatio="none"
        >
          <line
            x1={padding}
            y1={height - padding}
            x2={width - padding}
            y2={height - padding}
            stroke="#e2e8f0"
          />
          <polyline
            points={points}
            fill="none"
            stroke="#4f46e5"
            strokeWidth="3"
            strokeLinejoin="round"
            strokeLinecap="round"
          />
        </svg>
      )}
    </div>
  );
}

function Breakdown({
  title,
  items,
}: {
  title: string;
  items: LinkAnalytics["referrers"];
}) {
  return (
    <div>
      <h4 className="text-sm font-medium text-slate-700">{title}</h4>
      {items.length === 0 ? (
        <p className="mt-2 text-xs text-slate-500">No data</p>
      ) : (
        <ul className="mt-2 space-y-1">
          {items.map((item) => (
            <li
              key={item.name}
              className="flex justify-between gap-2 text-xs text-slate-600"
            >
              <span className="truncate">{item.name}</span>
              <span className="shrink-0 tabular-nums">{item.count}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function formatExpiration(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}
