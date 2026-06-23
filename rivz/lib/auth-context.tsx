"use client";
import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
} from "react";
import { api } from "./api";

type User = {
  id: string;
  email: string;
  role: string;
  display_name?: string | null;
  avatar_url?: string | null;
};
type AuthCtx = {
  user: User | null;
  token: string | null;
  login: (token: string, user: User) => void;
  logout: () => void;
  loading: boolean;
  refreshUser: () => Promise<void>;
};

const AuthContext = createContext<AuthCtx | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const login = useCallback((t: string, u: User) => {
    localStorage.setItem("token", t);
    setToken(t);
    setUser(u);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem("token");
    setToken(null);
    setUser(null);
  }, []);

  const refreshUser = useCallback(async () => {
    const stored = localStorage.getItem("token");
    if (!stored) return;
    try {
      const u = await api.get<{
        id: string;
        email: string;
        role: string;
        display_name?: string | null;
        avatar_url?: string | null;
      }>("/auth/me");
      setUser({
        id: u.id,
        email: u.email,
        role: u.role ?? "user",
        display_name: u.display_name,
        avatar_url: u.avatar_url,
      });
    } catch (e) {
      console.error("Failed to refresh user profile", e);
    }
  }, []);

  useEffect(() => {
    const stored = localStorage.getItem("token");
    if (!stored) {
      setLoading(false);
      return;
    }
    
    api.get<{
      id: string;
      email: string;
      role: string;
      display_name?: string | null;
      avatar_url?: string | null;
    }>("/auth/me")
      .then((u) => {
        setToken(stored);
        setUser({
          id: u.id,
          email: u.email,
          role: u.role ?? "user",
          display_name: u.display_name,
          avatar_url: u.avatar_url,
        });
      })
      .catch(() => {
        localStorage.removeItem("token");
        setToken(null);
        setUser(null);
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <AuthContext.Provider value={{ user, token, login, logout, loading, refreshUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
