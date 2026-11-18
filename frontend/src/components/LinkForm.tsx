import { useState, type FormEvent } from "react";
import { ApiError, createLink } from "../lib/api";
import type { Link } from "../types/link";

interface Props {
  onCreated: (link: Link) => void;
}

export default function LinkForm({ onCreated }: Props) {
  const [url, setUrl] = useState("");
  const [expiresAt, setExpiresAt] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const expires = expiresAt ? new Date(expiresAt).toISOString() : undefined;
      const link = await createLink(url.trim(), expires);
      onCreated(link);
      setUrl("");
      setExpiresAt("");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Terjadi kesalahan");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <input
        type="url"
        required
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        placeholder="https://contoh.com/url-yang-panjang"
        className="w-full rounded-lg border border-slate-300 px-4 py-3 outline-none focus:border-indigo-500"
      />
      <div className="flex flex-wrap items-center gap-3">
        <label className="text-sm text-slate-600">
          Kedaluwarsa (opsional){" "}
          <input
            type="datetime-local"
            value={expiresAt}
            onChange={(e) => setExpiresAt(e.target.value)}
            className="ml-1 rounded border border-slate-300 px-2 py-1"
          />
        </label>
        <button
          type="submit"
          disabled={loading}
          className="ml-auto rounded-lg bg-indigo-600 px-5 py-2.5 font-medium text-white hover:bg-indigo-700 disabled:opacity-60"
        >
          {loading ? "Memproses..." : "Perpendek"}
        </button>
      </div>
      {error && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}
    </form>
  );
}
