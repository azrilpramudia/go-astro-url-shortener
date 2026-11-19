import { useState, type SubmitEvent } from "react";
import { ApiError, login, register } from "../lib/api";

interface Props {
  mode: "login" | "register";
}

const inputClass =
  "min-h-12 w-full rounded-xl border border-slate-200 bg-white px-4 py-3 text-slate-900 shadow-sm outline-none transition placeholder:text-slate-400 hover:border-slate-300 focus-visible:border-indigo-500 focus-visible:ring-4 focus-visible:ring-indigo-100";

export default function AuthForm({ mode }: Props) {
  const isRegister = mode === "register";
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      if (isRegister) {
        await register(name.trim(), email.trim(), password);
      } else {
        await login(email.trim(), password);
      }
      window.location.href = "/";
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : "Something went wrong. Please try again.",
      );
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {isRegister && (
        <div className="space-y-2">
          <label
            htmlFor="auth-name"
            className="block text-sm font-medium text-slate-700"
          >
            Full name
          </label>
          <input
            id="auth-name"
            type="text"
            required
            minLength={2}
            maxLength={100}
            autoComplete="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Your name"
            className={inputClass}
          />
        </div>
      )}
      <div className="space-y-2">
        <label
          htmlFor="auth-email"
          className="block text-sm font-medium text-slate-700"
        >
          Email address
        </label>
        <input
          id="auth-email"
          type="email"
          required
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="you@example.com"
          className={inputClass}
        />
      </div>
      <div className="space-y-2">
        <label
          htmlFor="auth-password"
          className="block text-sm font-medium text-slate-700"
        >
          Password
        </label>
        <input
          id="auth-password"
          type="password"
          required
          minLength={isRegister ? 8 : undefined}
          maxLength={72}
          autoComplete={isRegister ? "new-password" : "current-password"}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Enter your password"
          aria-describedby={isRegister ? "password-hint" : undefined}
          className={inputClass}
        />
        {isRegister && (
          <p id="password-hint" className="text-xs text-slate-500">
            Use at least 8 characters.
          </p>
        )}
      </div>
      <button
        type="submit"
        disabled={loading}
        className="min-h-12 w-full rounded-xl bg-indigo-600 px-5 py-3 font-semibold text-white shadow-sm transition hover:bg-indigo-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:cursor-not-allowed disabled:opacity-60"
      >
        {loading
          ? isRegister
            ? "Creating account..."
            : "Signing in..."
          : isRegister
            ? "Create account"
            : "Sign in"}
      </button>
      {error && (
        <p
          role="alert"
          className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
        >
          {error}
        </p>
      )}
      <p className="text-center text-sm text-slate-600">
        {isRegister ? (
          <>
            Already have an account?{" "}
            <a
              href="/login/"
              className="font-semibold text-indigo-700 hover:underline focus-visible:rounded-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
              Sign in
            </a>
          </>
        ) : (
          <>
            New to URL Shortener?{" "}
            <a
              href="/register/"
              className="font-semibold text-indigo-700 hover:underline focus-visible:rounded-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
              Create an account
            </a>
          </>
        )}
      </p>
    </form>
  );
}
