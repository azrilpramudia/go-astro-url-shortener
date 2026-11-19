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
      <p className="text-slate-500">
        No links yet. Create your first one above.
      </p>
    );
  }

  return (
    <ul className="divide-y divide-slate-200 rounded-lg border border-slate-200">
      {links.map((link) => {
        const expiration = link.expires_at
          ? formatExpiration(link.expires_at)
          : null;
        const feedback =
          copyFeedback?.code === link.code ? copyFeedback.status : null;

        return (
          <li
            key={link.code}
            className="flex flex-col gap-3 p-3 sm:flex-row sm:items-center sm:p-4"
          >
            <div className="min-w-0 flex-1">
              <a
                href={link.short_url}
                target="_blank"
                rel="noopener noreferrer"
                className="block break-all font-medium text-indigo-600 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:break-normal sm:truncate"
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
                  className="inline-flex min-h-11 flex-1 items-center justify-center rounded border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:flex-none"
                >
                  {feedback === "copied" ? "Copied" : "Copy"}
                </button>
                {onRemove && (
                  <button
                    onClick={() => onRemove(link.code)}
                    aria-label={`Hapus ${link.code}`}
                    className="text-sm text-slate-400 hover:text-red-600"
                  >
                    Hapus
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
