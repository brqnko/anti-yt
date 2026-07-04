import { createContext, type ComponentChildren } from "preact";
import { useState, useEffect, useCallback, useContext, useMemo } from "preact/hooks";
import { AxiosError } from "axios";
import { getUser } from "../api/generated/user";
import { getAuth } from "../api/generated/auth";
import { getCookie } from "../utils/cookie";

type ScreenTimeBlockReason = "limit_exceeded" | "outside_time_range";

interface AuthState {
  isLoading: boolean;
  isAuthenticated: boolean;
  error: Error | null;
  screenTimeBlocked: boolean;
  screenTimeBlockReason: ScreenTimeBlockReason | null;
  logout: () => Promise<void>;
  refreshAuth: () => Promise<void>;
  clearScreenTimeBlock: () => void;
}

const AuthContext = createContext<AuthState>({
  isLoading: true,
  isAuthenticated: false,
  error: null,
  screenTimeBlocked: false,
  screenTimeBlockReason: null,
  logout: async () => {},
  refreshAuth: async () => {},
  clearScreenTimeBlock: () => {},
});

export function AuthProvider({ children }: { children: ComponentChildren }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [screenTimeBlocked, setScreenTimeBlocked] = useState(false);
  const [screenTimeBlockReason, setScreenTimeBlockReason] = useState<ScreenTimeBlockReason | null>(null);

  const checkAuth = useCallback(async () => {
    if (typeof window === "undefined") {
      setIsLoading(false);
      return;
    }

    // Skip the status request entirely when there's no CSRF cookie. The API
    // would respond 401 (which would then trigger a refresh that 400s),
    // surfacing console errors even though the app already handles the case.
    if (!getCookie("csrf_token")) {
      setIsAuthenticated(false);
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
      if (err instanceof AxiosError && err.response?.status === 403) {
        const title = err.response?.data?.title;
        if (title === "screen_time_limit_exceeded" || title === "outside_allowed_time_range") {
          setIsAuthenticated(true);
          return;
        }
      }
      setIsAuthenticated(false);
      if (!(err instanceof AxiosError) || err.response?.status !== 401) {
        setError(
          err instanceof Error ? err : new Error("Failed to check auth"),
        );
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  const clearScreenTimeBlock = useCallback(() => {
    setScreenTimeBlocked(false);
    setScreenTimeBlockReason(null);
  }, []);

  const logout = useCallback(async () => {
    try {
      const { postAuthLogout } = getAuth();
      await postAuthLogout();
    } catch {
    }
    setIsAuthenticated(false);
    window.location.href = "/";
  }, []);

  useEffect(() => {
    const handler = () => {
      setIsAuthenticated(false);
      setIsLoading(false);
    };
    window.addEventListener("auth:logout", handler);
    return () => window.removeEventListener("auth:logout", handler);
  }, []);

  useEffect(() => {
    const handler = (e: Event) => {
      if (!(e instanceof CustomEvent)) return;
      const reason = e.detail?.reason;
      setScreenTimeBlocked(true);
      setScreenTimeBlockReason(
        reason === "outside_time_range" ? "outside_time_range" : "limit_exceeded",
      );
    };
    window.addEventListener("screen-time:blocked", handler);
    return () => window.removeEventListener("screen-time:blocked", handler);
  }, []);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  const value = useMemo(
    () => ({
      isLoading,
      isAuthenticated,
      error,
      screenTimeBlocked,
      screenTimeBlockReason,
      logout,
      refreshAuth: checkAuth,
      clearScreenTimeBlock,
    }),
    [
      isLoading,
      isAuthenticated,
      error,
      screenTimeBlocked,
      screenTimeBlockReason,
      logout,
      checkAuth,
      clearScreenTimeBlock,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  return useContext(AuthContext);
}
