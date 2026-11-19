import { useState, type FormEvent } from "react";
import { ApiError, login, register } from "../lib/api";

interface Props {
  mode: "login" | "register";
}

const inputClass =
  "w-full rounded-lg border border-slate-300 px-4 py-3 outline-none focus:border-indigo-500";

export default function AuthForm({ mode }: Props) {
  const isRegister = mode === "register";
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: FormEvent) {
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
      setError(err instanceof ApiError ? err.message : "Terjadi kesalahan");
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      {isRegister && (
        <input
          type="text"
          required
          minLength={2}
          maxLength={100}
          autoComplete="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Nama"
          className={inputClass}
        />
      )}
      <input
        type="email"
        required
        autoComplete="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
        className={inputClass}
      />
      <input
        type="password"
        required
        minLength={isRegister ? 8 : undefined}
        maxLength={72}
        autoComplete={isRegister ? "new-password" : "current-password"}
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder={isRegister ? "Password (minimal 8 karakter)" : "Password"}
        className={inputClass}
      />
      <button
        type="submit"
        disabled={loading}
        className="w-full rounded-lg bg-indigo-600 px-5 py-3 font-medium text-white hover:bg-indigo-700 disabled:opacity-60"
      >
        {loading ? "Memproses..." : isRegister ? "Daftar" : "Masuk"}
      </button>
      {error && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}
      <p className="text-sm text-slate-600">
        {isRegister ? (
          <>
            Sudah punya akun?{" "}
            <a href="/login" className="text-indigo-600 hover:underline">
              Masuk
            </a>
          </>
        ) : (
          <>
            Belum punya akun?{" "}
            <a href="/register" className="text-indigo-600 hover:underline">
              Daftar
            </a>
          </>
        )}
      </p>
    </form>
  );
}
