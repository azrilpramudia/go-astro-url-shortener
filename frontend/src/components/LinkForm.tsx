import { useState, type FormEvent } from "react";
import { ApiError, createLink } from "../lib/api";
import { copyToClipboard } from "../lib/clipboard";
import type { Link } from "../types/link";

interface Props {
  onCreated: (link: Link) => void;
}

export default function LinkForm({ onCreated }: Props) {
  const [url, setUrl] = useState("");
  const [expiresAt, setExpiresAt] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdLink, setCreatedLink] = useState<Link | null>(null);
  const [copyStatus, setCopyStatus] = useState<"copied" | "error" | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setCreatedLink(null);
    setCopyStatus(null);
    try {
      const expires = expiresAt ? new Date(expiresAt).toISOString() : undefined;
      const link = await createLink(url.trim(), expires);
      onCreated(link);
      setCreatedLink(link);
      setUrl("");
      setExpiresAt("");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong.");
    } finally {
      setLoading(false);
    }
  }

  async function copyCreatedLink() {
    if (!createdLink) return;

    try {
      await copyToClipboard(createdLink.short_url);
      setCopyStatus("copied");
    } catch {
      setCopyStatus("error");
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      <label
        htmlFor="long-url"
        className="block text-sm font-medium text-slate-700"
      >
        Long URL
      </label>
      <input
        id="long-url"
        type="url"
        required
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        placeholder="https://example.com/a-very-long-url"
        className="min-h-12 w-full min-w-0 rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-slate-900 shadow-sm outline-none transition placeholder:text-slate-400 hover:border-slate-300 focus-visible:border-indigo-500 focus-visible:bg-white focus-visible:ring-4 focus-visible:ring-indigo-100"
      />
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end">
        <label className="block min-w-0 flex-1 text-sm font-medium text-slate-700">
          Expires at <span className="font-normal text-slate-500">(optional)</span>
          <input
            type="datetime-local"
            value={expiresAt}
            onChange={(e) => setExpiresAt(e.target.value)}
            className="mt-2 block min-h-11 w-full min-w-0 rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 shadow-sm outline-none transition focus-visible:border-indigo-500 focus-visible:ring-4 focus-visible:ring-indigo-100 sm:max-w-xs"
          />
        </label>
        <button
          type="submit"
          disabled={loading}
          className="min-h-12 w-full rounded-xl bg-indigo-600 px-6 py-3 font-semibold text-white shadow-sm transition hover:bg-indigo-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto"
        >
          {loading ? "Shortening..." : "Shorten"}
        </button>
      </div>
      {error && (
        <p
          role="alert"
          className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
        >
          {error}
        </p>
      )}
      {createdLink && (
        <section className="space-y-3 rounded-xl border border-emerald-200 bg-emerald-50 p-4 sm:p-5">
          <p role="status" className="font-semibold text-emerald-900">
            Short link created successfully.
          </p>
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <a
              href={createdLink.short_url}
              target="_blank"
              rel="noopener noreferrer"
              className="min-w-0 break-all font-semibold text-indigo-700 underline decoration-indigo-300 underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
              {createdLink.short_url}
            </a>
            <button
              type="button"
              onClick={copyCreatedLink}
              className="min-h-11 shrink-0 rounded-lg border border-emerald-300 bg-white px-4 py-2 text-sm font-semibold text-emerald-900 shadow-sm transition hover:bg-emerald-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
              {copyStatus === "copied" ? "Copied" : "Copy short link"}
            </button>
          </div>
          {copyStatus === "copied" && (
            <p role="status" className="text-sm text-emerald-800">
              Copied to clipboard.
            </p>
          )}
          {copyStatus === "error" && (
            <p role="alert" className="text-sm text-red-700">
              Could not copy the link. Please copy it manually.
            </p>
          )}
        </section>
      )}
    </form>
  );
}
