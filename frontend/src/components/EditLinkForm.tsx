import { useState, type SubmitEvent } from "react";
import { ApiError, updateLink } from "../lib/api";
import type { Link } from "../types/link";

interface Props {
  link: Link;
  onUpdated: (link: Link) => void;
  onCancel: () => void;
}

export default function EditLinkForm({ link, onUpdated, onCancel }: Props) {
  const [url, setUrl] = useState(link.original_url);
  const [expiresAt, setExpiresAt] = useState(toDateTimeLocal(link.expires_at));
  const [tags, setTags] = useState((link.tags ?? []).join(", "));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const updated = await updateLink(link.code, {
        url: url.trim(),
        expires_at: expiresAt ? new Date(expiresAt).toISOString() : undefined,
        tags: tags.split(","),
      });
      onUpdated(updated);
    } catch (cause) {
      setError(
        cause instanceof ApiError
          ? cause.message
          : "Could not update this link. Please try again.",
      );
    } finally {
      setSaving(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      aria-label={`Edit link ${link.code}`}
      className="mt-3 space-y-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
    >
      <label className="block text-sm font-medium text-slate-700">
        Destination URL
        <input
          type="url"
          required
          maxLength={2048}
          value={url}
          onChange={(event) => setUrl(event.target.value)}
          className="mt-1 min-h-10 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
        />
      </label>
      <label className="block text-sm font-medium text-slate-700">
        Expires at
        <input
          type="datetime-local"
          value={expiresAt}
          onChange={(event) => setExpiresAt(event.target.value)}
          className="mt-1 min-h-10 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
        />
      </label>
      <label className="block text-sm font-medium text-slate-700">
        Tags <span className="font-normal text-slate-500">(comma-separated)</span>
        <input
          type="text"
          value={tags}
          onChange={(event) => setTags(event.target.value)}
          placeholder="campaign, product"
          className="mt-1 min-h-10 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
        />
      </label>
      {error && (
        <p role="alert" className="text-sm text-red-700">
          {error}
        </p>
      )}
      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onCancel}
          className="min-h-10 rounded-lg px-3 text-sm font-medium text-slate-600 hover:bg-slate-100"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={saving}
          className="min-h-10 rounded-lg bg-indigo-600 px-4 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
        >
          {saving ? "Saving..." : "Save changes"}
        </button>
      </div>
    </form>
  );
}

function toDateTimeLocal(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const localDate = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return localDate.toISOString().slice(0, 16);
}
