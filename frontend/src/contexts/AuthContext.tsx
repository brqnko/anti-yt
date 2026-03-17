import { createContext } from "preact";
import { useState, useEffect, useCallback, useContext } from "preact/hooks";
import type { ComponentChildren } from "preact";
import { getUser } from "../api/generated/user";
import { getAuth } from "../api/generated/auth";

interface AuthState {
  isLoading: boolean;
  isAuthenticated: boolean;
  error: Error | null;
  sessionExpired: boolean;
  screenTimeBlocked: boolean;
  screenTimeBlockReason: "limit_exceeded" | "outside_time_range" | null;
  logout: () => Promise<void>;
  refreshAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthState>({
  isLoading: true,
  isAuthenticated: false,
  error: null,
  sessionExpired: false,
  screenTimeBlocked: false,
  screenTimeBlockReason: null,
  logout: async () => {},
  refreshAuth: async () => {},
});

export function AuthProvider({ children }: { children: ComponentChildren }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [sessionExpired, setSessionExpired] = useState(false);
  const [screenTimeBlocked, setScreenTimeBlocked] = useState(false);
  const [screenTimeBlockReason, setScreenTimeBlockReason] = useState<
    "limit_exceeded" | "outside_time_range" | null
  >(null);

  const checkAuth = useCallback(async () => {
    if (typeof window === "undefined") {
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      const { getUsersMeStatus } = getUser();
      await getUsersMeStatus();
      setIsAuthenticated(true);
    } catch (err: unknown) {
      setIsAuthenticated(false);
      const axiosErr = err as { response?: { status?: number } };
      if (axiosErr.response?.status !== 401) {
        setError(
          err instanceof Error ? err : new Error("Failed to check auth"),
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
    setIsAuthenticated(false);
    window.location.href = "/";
  }, []);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent).detail;
      setIsAuthenticated(false);
      setIsLoading(false);
      if (detail?.reason === "session_expired") {
        setSessionExpired(true);
      }
    };
    window.addEventListener("auth:logout", handler);
    return () => window.removeEventListener("auth:logout", handler);
  }, []);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent).detail;
      setScreenTimeBlocked(true);
      setScreenTimeBlockReason(detail?.reason ?? "limit_exceeded");
    };
    window.addEventListener("screen-time:blocked", handler);
    return () => window.removeEventListener("screen-time:blocked", handler);
  }, []);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  return (
    <AuthContext.Provider
      value={{
        isLoading,
        isAuthenticated,
        error,
        sessionExpired,
        screenTimeBlocked,
        screenTimeBlockReason,
        logout,
        refreshAuth: checkAuth,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthState {
  return useContext(AuthContext);
}
