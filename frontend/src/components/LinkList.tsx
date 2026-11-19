import { useState } from "react";
import { copyToClipboard } from "../lib/clipboard";
import type { Link } from "../types/link";

interface Props {
  links: Link[];
  onRemove?: (code: string) => void;
}

export default function LinkList({ links, onRemove }: Props) {
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
            </div>
            <div className="flex flex-col gap-2 sm:items-end">
              <div className="flex flex-wrap items-center gap-2 sm:flex-nowrap">
                <span className="mr-auto whitespace-nowrap text-sm text-slate-600 sm:mr-0">
                  {link.clicks} clicks
                </span>
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

function formatExpiration(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}
