import { useCallback, useEffect, useState } from "react";
import { getLink } from "../lib/api";
import { addCode, loadCodes, removeCode } from "../lib/storage";
import type { Link } from "../types/link";
import LinkForm from "./LinkForm";
import LinkList from "./LinkList";

export default function LinkManager() {
  const [links, setLinks] = useState<Link[]>([]);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    const results = await Promise.allSettled(loadCodes().map(getLink));
    setLinks(
      results.flatMap((r) => (r.status === "fulfilled" ? [r.value] : [])),
    );
    setLoading(false);
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
    <div className="space-y-10">
      <LinkForm onCreated={handleCreated} />

      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Link kamu</h2>
          <button
            onClick={refresh}
            disabled={loading}
            className="text-sm text-indigo-600 hover:underline disabled:opacity-60"
          >
            {loading ? "Memuat..." : "Muat ulang statistik"}
          </button>
        </div>
        <LinkList links={links} onRemove={handleRemove} />
      </section>
    </div>
  );
}
