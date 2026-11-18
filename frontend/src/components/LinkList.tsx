import { useState } from "react";
import type { Link } from "../types/link";

interface Props {
  links: Link[];
  onRemove: (code: string) => void;
}

export default function LinkList({ links, onRemove }: Props) {
  const [copied, setCopied] = useState<string | null>(null);

  async function copy(link: Link) {
    try {
      await navigator.clipboard.writeText(link.short_url);
      setCopied(link.code);
      setTimeout(() => setCopied(null), 1500);
    } catch {}
  }

  if (links.length === 0) {
    return (
      <p className="text-slate-500">
        Belum ada link. Buat yang pertama di atas.
      </p>
    );
  }

  return (
    <ul className="divide-y divide-slate-200 rounded-lg border border-slate-200">
      {links.map((link) => (
        <li key={link.code} className="flex items-center gap-4 p-4">
          <div className="min-w-0 flex-1">
            <a
              href={link.short_url}
              target="_blank"
              rel="noopener noreferrer"
              className="font-medium text-indigo-600 hover:underline"
            >
              {link.short_url}
            </a>
            <p className="truncate text-sm text-slate-500">
              {link.original_url}
            </p>
          </div>
          <span className="whitespace-nowrap text-sm text-slate-600">
            {link.clicks} klik
          </span>
          <button
            onClick={() => copy(link)}
            className="rounded border border-slate-300 px-3 py-1 text-sm hover:bg-slate-50"
          >
            {copied === link.code ? "Tersalin" : "Salin"}
          </button>
          <button
            onClick={() => onRemove(link.code)}
            aria-label={`Hapus ${link.code} dari daftar`}
            className="text-sm text-slate-400 hover:text-red-600"
          >
            Hapus
          </button>
        </li>
      ))}
    </ul>
  );
}
