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
          setError(
            err instanceof ApiError
              ? err.message
              : "Something went wrong. Please try again.",
          );
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
      setError(
        err instanceof ApiError
          ? err.message
          : "Something went wrong. Please try again.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function handleRemove(code: string) {
    if (!window.confirm("Permanently delete this link?")) return;
    setError(null);
    try {
      await deleteLink(code);
      setLinks((prev) => prev.filter((l) => l.code !== code));
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : "Something went wrong. Please try again.",
      );
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
    return (
      <p role="status" className="py-12 text-center text-sm text-slate-500">
        Loading your workspace...
      </p>
    );
  }

  return (
    <div className="space-y-10 sm:space-y-12">
      <header className="flex items-center justify-end gap-3">
        {user ? (
          <nav aria-label="Account" className="flex items-center gap-2 sm:gap-4">
            <span className="hidden max-w-40 truncate text-sm text-slate-600 sm:inline">
              Hi, {user.name}
            </span>
            <button
              onClick={handleLogout}
              className="min-h-10 rounded-lg px-3 text-sm font-medium text-slate-600 transition hover:bg-slate-100 hover:text-slate-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
              Sign out
            </button>
          </nav>
        ) : (
          <nav aria-label="Account" className="flex items-center gap-2 sm:gap-3">
            <a
              href="/login/"
              className="inline-flex min-h-10 items-center rounded-lg px-3 text-sm font-medium text-slate-600 transition hover:bg-slate-100 hover:text-slate-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
              Sign in
            </a>
            <a
              href="/register/"
              className="inline-flex min-h-10 items-center rounded-lg bg-indigo-600 px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-indigo-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
              Sign up
            </a>
          </nav>
        )}
      </header>

      <section
        aria-label="Create a short link"
        className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm shadow-slate-200/50 sm:p-7"
      >
        <div className="mb-5">
          <h2 className="text-lg font-semibold tracking-tight text-slate-900">
            Create a short link
          </h2>
          <p className="mt-1 text-sm text-slate-500">
            Paste a long URL to get a shorter, shareable link.
          </p>
        </div>
        <LinkForm onCreated={(link) => setLinks((prev) => [link, ...prev])} />
      </section>

      {error && (
        <p
          role="alert"
          className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
        >
          {error}
        </p>
      )}

      <section aria-labelledby="links-heading" className="space-y-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2
              id="links-heading"
              className="text-lg font-semibold tracking-tight text-slate-900"
            >
              {user ? "Your links" : "Recent links"}
            </h2>
            <p className="mt-1 text-sm text-slate-500">
              {user
                ? "View and manage your saved short links."
                : "Links you create will appear here."}
            </p>
          </div>
          {user && (
            <button
              onClick={refresh}
              disabled={busy}
              className="min-h-10 self-start rounded-lg px-3 text-sm font-medium text-indigo-700 transition hover:bg-indigo-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:cursor-not-allowed disabled:opacity-60 sm:self-auto"
            >
              {busy ? "Refreshing..." : "Refresh"}
            </button>
          )}
        </div>

        <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm shadow-slate-200/40">
          <LinkList
            links={links}
            onRemove={user ? handleRemove : undefined}
            analyticsEnabled={Boolean(user)}
          />
        </div>

        {!user && links.length > 0 && (
          <p className="text-sm leading-6 text-slate-500">
            Links created while signed out are available only in this session.{" "}
            <a
              href="/login/"
              className="font-medium text-indigo-700 hover:underline"
            >
              Sign in
            </a>{" "}
            or{" "}
            <a
              href="/register/"
              className="font-medium text-indigo-700 hover:underline"
            >
              create an account
            </a>{" "}
            to save and manage your links.
          </p>
        )}
      </section>
    </div>
  );
}
