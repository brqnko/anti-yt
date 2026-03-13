import { createContext } from "preact";
import { useState, useEffect, useCallback, useContext } from "preact/hooks";
import type { ComponentChildren } from "preact";
import type { User } from "../api/generated/antiYtApi.schemas";
import { getUser } from "../api/generated/user";
import { getAuth } from "../api/generated/auth";

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  error: Error | null;
  sessionExpired: boolean;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthState>({
  user: null,
  isLoading: true,
  isAuthenticated: false,
  error: null,
  sessionExpired: false,
  logout: async () => {},
  refreshUser: async () => {},
});

export function AuthProvider({ children }: { children: ComponentChildren }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [sessionExpired, setSessionExpired] = useState(false);

  const fetchUser = useCallback(async () => {
    if (typeof window === "undefined") {
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      const { getUsersMeStatus } = getUser();
      const userData = await getUsersMeStatus();
      setUser(userData);
    } catch (err: unknown) {
      setUser(null);
      const axiosErr = err as { response?: { status?: number } };
      if (axiosErr.response?.status !== 401) {
        setError(
          err instanceof Error ? err : new Error("Failed to fetch user"),
        );
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      const { postAuthLogout } = getAuth();
      await postAuthLogout();
    } catch {
      // Clear state regardless
    }
    setUser(null);
    window.location.href = "/";
  }, []);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent).detail;
      setUser(null);
      setIsLoading(false);
      if (detail?.reason === "session_expired") {
        setSessionExpired(true);
      }
    };
    window.addEventListener("auth:logout", handler);
    return () => window.removeEventListener("auth:logout", handler);
  }, []);

  useEffect(() => {
    fetchUser();
  }, [fetchUser]);

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: user !== null,
        error,
        sessionExpired,
        logout,
        refreshUser: fetchUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthState {
  return useContext(AuthContext);
}
