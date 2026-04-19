import { useState, useEffect } from "react";
import { login, register } from "@/lib/graphql/client";

interface AuthModalProps {
  onClose: () => void;
  onSuccess: (name: string) => void;
}

export function AuthModal({ onClose, onSuccess }: AuthModalProps) {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fullName, setFullName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleEscape);
    return () => window.removeEventListener("keydown", handleEscape);
  }, [onClose]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      if (isLogin) {
        const payload = await login(email, password);
        localStorage.setItem("token", payload.token);
        localStorage.setItem("externalKey", payload.externalKey);
        onSuccess(payload.studentName);
      } else {
        const payload = await register(email, fullName, password);
        localStorage.setItem("token", payload.token);
        localStorage.setItem("externalKey", payload.externalKey);
        onSuccess(payload.studentName);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Authentication failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-ink/20 p-4 backdrop-blur-sm">
      <div className="relative w-full max-w-sm rounded-3xl bg-white p-8 shadow-2xl">
        <button
          onClick={onClose}
          className="absolute right-4 top-4 text-ink/40 hover:text-ink"
        >
          <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>

        <h2 className="mb-6 font-display text-2xl font-semibold text-ink">
          {isLogin ? "Welcome back" : "Create account"}
        </h2>

        {error && (
          <div className="mb-4 rounded-xl bg-rose/10 p-3 text-sm text-rose">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          {!isLogin && (
            <div>
              <label className="mb-1 block text-sm font-medium text-ink/70">Full Name</label>
              <input
                type="text"
                required
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                className="w-full rounded-xl border border-ink/10 bg-cloud px-4 py-3 text-ink outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20"
                placeholder="John Doe"
              />
            </div>
          )}

          <div>
            <label className="mb-1 block text-sm font-medium text-ink/70">Email</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full rounded-xl border border-ink/10 bg-cloud px-4 py-3 text-ink outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20"
              placeholder="you@uwaterloo.ca"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-ink/70">Password</label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rounded-xl border border-ink/10 bg-cloud px-4 py-3 text-ink outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="mt-6 w-full rounded-xl bg-teal py-3 text-center font-semibold text-white transition hover:bg-teal/90 disabled:opacity-70"
          >
            {loading ? "Please wait..." : isLogin ? "Sign In" : "Sign Up"}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-ink/60">
          {isLogin ? "Don't have an account? " : "Already have an account? "}
          <button
            type="button"
            className="font-medium text-teal hover:underline"
            onClick={() => setIsLogin(!isLogin)}
          >
            {isLogin ? "Sign up" : "Log in"}
          </button>
        </p>
      </div>
    </div>
  );
}
