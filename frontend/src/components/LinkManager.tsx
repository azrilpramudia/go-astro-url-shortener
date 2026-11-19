import { useEffect, useState } from "react";
import { ApiError, deleteLink, getMe, listMyLinks, logout } from "../lib/api";
import type { Link, User } from "../types/link";
import LinkForm from "./LinkForm";
import LinkList from "./LinkList";

export default function LinkManager() {
  const [user, setUser] = useState<User | null | undefined>(undefined);
  const [links, setLinks] = useState<Link[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function init() {
      try {
        const me = await getMe();
        const page = await listMyLinks();
        if (!cancelled) {
          setUser(me);
          setLinks(page.items);
        }
      } catch (err) {
        if (cancelled) return;
        setUser(null);
        if (!(err instanceof ApiError && err.status === 401)) {
          setError(err instanceof ApiError ? err.message : "Terjadi kesalahan");
        }
      }
    }

    init();
    return () => {
      cancelled = true;
    };
  }, []);

  async function refresh() {
    setBusy(true);
    setError(null);
    try {
      setLinks((await listMyLinks()).items);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Terjadi kesalahan");
    } finally {
      setBusy(false);
    }
  }

  async function handleRemove(code: string) {
    if (!window.confirm("Hapus link ini secara permanen?")) return;
    setError(null);
    try {
      await deleteLink(code);
      setLinks((prev) => prev.filter((l) => l.code !== code));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Terjadi kesalahan");
    }
  }

  async function handleLogout() {
    try {
      await logout();
    } finally {
      setUser(null);
      setLinks([]);
    }
  }

  if (user === undefined) {
    return <p className="text-slate-500">Memuat...</p>;
  }

  return (
    <div className="space-y-10">
      <div className="flex items-center justify-end gap-4 text-sm">
        {user ? (
          <>
            <span className="text-slate-600">Halo, {user.name}</span>
            <button
              onClick={handleLogout}
              className="text-indigo-600 hover:underline"
            >
              Keluar
            </button>
          </>
        ) : (
          <>
            <a href="/login" className="text-indigo-600 hover:underline">
              Masuk
            </a>
            <a
              href="/register"
              className="rounded bg-indigo-600 px-3 py-1.5 text-white hover:bg-indigo-700"
            >
              Daftar
            </a>
          </>
        )}
      </div>

      <LinkForm onCreated={(link) => setLinks((prev) => [link, ...prev])} />

      {error && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}

      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold">
            {user ? "Link kamu" : "Link yang baru dibuat"}
          </h2>
          {user && (
            <button
              onClick={refresh}
              disabled={busy}
              className="text-sm text-indigo-600 hover:underline disabled:opacity-60"
            >
              {busy ? "Memuat..." : "Muat ulang statistik"}
            </button>
          )}
        </div>

        <LinkList links={links} onRemove={user ? handleRemove : undefined} />

        {!user && links.length > 0 && (
          <p className="mt-3 text-sm text-slate-500">
            Link ini hanya tampil di halaman ini.{" "}
            <a href="/login" className="text-indigo-600 hover:underline">
              Masuk
            </a>{" "}
            atau{" "}
            <a href="/register" className="text-indigo-600 hover:underline">
              daftar
            </a>{" "}
            untuk menyimpannya di dashboard.
          </p>
        )}
      </section>
    </div>
  );
}
