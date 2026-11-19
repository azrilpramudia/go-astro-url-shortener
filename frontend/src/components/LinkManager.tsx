import { useEffect, useState } from "react";
import { ApiError, deleteLink, getMe, listMyLinks, logout } from "../lib/api";
import type { Link, LinkPage, User } from "../types/link";
import LinkForm from "./LinkForm";
import LinkList from "./LinkList";

const pageSize = 20;

export default function LinkManager() {
  const [user, setUser] = useState<User | null | undefined>(undefined);
  const [links, setLinks] = useState<Link[]>([]);
  const [linksLoaded, setLinksLoaded] = useState(false);
  const [pageInfo, setPageInfo] = useState<LinkPage>({
    items: [],
    total: 0,
    page: 1,
    size: pageSize,
  });
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [status, setStatus] = useState<"" | "active" | "expired">("");
  const [tag, setTag] = useState("");
  const [reloadKey, setReloadKey] = useState(0);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function init() {
      try {
        const me = await getMe();
        if (!cancelled) {
          setUser(me);
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

  useEffect(() => {
    const timer = window.setTimeout(
      () => setDebouncedQuery(query.trim()),
      250,
    );
    return () => window.clearTimeout(timer);
  }, [query]);

  useEffect(() => {
    if (!user) return;
    let cancelled = false;
    setBusy(true);
    setError(null);
    listMyLinks(page, pageSize, { q: debouncedQuery, status, tag })
      .then((result) => {
        if (!cancelled) {
          setLinks(result.items);
          setPageInfo(result);
          setLinksLoaded(true);
        }
      })
      .catch((cause) => {
        if (!cancelled) {
          setLinksLoaded(true);
          setError(
            cause instanceof ApiError
              ? cause.message
              : "Something went wrong. Please try again.",
          );
        }
      })
      .finally(() => {
        if (!cancelled) setBusy(false);
      });

    return () => {
      cancelled = true;
    };
  }, [user, page, debouncedQuery, status, tag, reloadKey]);

  async function refresh() {
    setReloadKey((key) => key + 1);
  }

  async function handleRemove(code: string) {
    if (!window.confirm("Permanently delete this link?")) return;
    setError(null);
    try {
      await deleteLink(code);
      if (links.length === 1 && page > 1) setPage((current) => current - 1);
      else setReloadKey((key) => key + 1);
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
      setLinksLoaded(false);
      setPageInfo({ items: [], total: 0, page: 1, size: pageSize });
    }
  }

  function handleCreated(link: Link) {
    if (user) {
      setPage(1);
      setReloadKey((key) => key + 1);
    } else {
      setLinks((previous) => [link, ...previous]);
    }
  }

  function handleUpdated(link: Link) {
    setLinks((previous) =>
      previous.map((current) => (current.code === link.code ? link : current)),
    );
    setReloadKey((key) => key + 1);
  }

  const pageCount = Math.max(1, Math.ceil(pageInfo.total / pageSize));

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
        <LinkForm onCreated={handleCreated} authenticated={Boolean(user)} />
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

        {user && (
          <div className="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 sm:grid-cols-[minmax(12rem,1fr)_10rem_minmax(10rem,0.7fr)]">
            <label className="text-sm font-medium text-slate-700">
              Search
              <input
                type="search"
                value={query}
                onChange={(event) => {
                  setPage(1);
                  setQuery(event.target.value);
                }}
                placeholder="Code or destination URL"
                className="mt-1 min-h-10 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
              />
            </label>
            <label className="text-sm font-medium text-slate-700">
              Status
              <select
                value={status}
                onChange={(event) => {
                  const nextStatus = event.target.value;
                  setPage(1);
                  if (
                    nextStatus === "" ||
                    nextStatus === "active" ||
                    nextStatus === "expired"
                  ) {
                    setStatus(nextStatus);
                  }
                }}
                className="mt-1 min-h-10 w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm"
              >
                <option value="">All links</option>
                <option value="active">Active</option>
                <option value="expired">Expired</option>
              </select>
            </label>
            <label className="text-sm font-medium text-slate-700">
              Tag
              <input
                type="search"
                value={tag}
                onChange={(event) => {
                  setPage(1);
                  setTag(event.target.value);
                }}
                placeholder="Filter by tag"
                className="mt-1 min-h-10 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
              />
            </label>
          </div>
        )}

        <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm shadow-slate-200/40">
          {user && !linksLoaded ? (
            <p role="status" className="px-5 py-10 text-center text-sm text-slate-500">
              Loading your links...
            </p>
          ) : (
            <LinkList
              links={links}
              onRemove={user ? handleRemove : undefined}
              analyticsEnabled={Boolean(user)}
              onUpdate={user ? handleUpdated : undefined}
            />
          )}
        </div>

        {user && (
          <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-slate-600">
            <span>
              {busy
                ? "Loading links..."
                : `${pageInfo.total.toLocaleString()} links · Page ${page} of ${pageCount}`}
            </span>
            {pageCount > 1 && (
              <div className="flex gap-2">
                <button
                  type="button"
                  disabled={page <= 1 || busy}
                  onClick={() => setPage((current) => current - 1)}
                  className="min-h-9 rounded-lg border border-slate-200 px-3 disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  type="button"
                  disabled={page >= pageCount || busy}
                  onClick={() => setPage((current) => current + 1)}
                  className="min-h-9 rounded-lg border border-slate-200 px-3 disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            )}
          </div>
        )}

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
