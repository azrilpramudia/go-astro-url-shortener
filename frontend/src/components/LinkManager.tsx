import { useCallback, useEffect, useState } from "react";
import { getLink } from "../lib/api";
import { addCode, loadCodes, removeCode } from "../lib/storage";
import type { Link } from "../types/link";
import LinkForm from "./LinkForm";
import LinkList from "./LinkList";

export default function LinkManager() {
  const [links, setLinks] = useState<Link[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    setLoadError(false);
    try {
      const results = await Promise.allSettled(loadCodes().map(getLink));
      setLinks(
        results.flatMap((r) => (r.status === "fulfilled" ? [r.value] : [])),
      );
      setLoadError(results.some((r) => r.status === "rejected"));
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  function handleCreated(link: Link) {
    addCode(link.code);
    setLinks((prev) => [link, ...prev]);
  }

  function handleRemove(code: string) {
    removeCode(code);
    setLinks((prev) => prev.filter((l) => l.code !== code));
  }

  return (
    <div className="space-y-8 sm:space-y-10">
      <LinkForm onCreated={handleCreated} />

      <section>
        <div className="mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <h2 className="text-lg font-semibold">Your links</h2>
          <button
            onClick={refresh}
            disabled={loading}
            className="min-h-11 self-start text-left text-sm text-indigo-600 hover:underline disabled:opacity-60 sm:self-auto sm:text-right"
          >
            {loading ? "Loading..." : "Refresh statistics"}
          </button>
        </div>
        <p className="mb-3 text-sm text-slate-500">
          Your link list is stored in this browser only and won&apos;t sync across devices.
        </p>
        {loadError && (
          <p role="alert" className="mb-3 text-sm text-red-600">
            {links.length > 0
              ? "Some saved links could not be loaded. The list may be incomplete. Try refreshing."
              : "Unable to load your saved links. Please try refreshing."}
          </p>
        )}
        {(links.length > 0 || !loadError) && (
          <LinkList links={links} onRemove={handleRemove} />
        )}
      </section>
    </div>
  );
}
