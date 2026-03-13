import { useLocation } from "preact-iso";
import { useEffect } from "preact/hooks";
import type { ComponentChildren } from "preact";
import { useAuth } from "../contexts/AuthContext";

interface ProtectedRouteProps {
  children: ComponentChildren;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, sessionExpired } = useAuth();
  const { route } = useLocation();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      route(sessionExpired ? "/?expired=1" : "/");
    }
  }, [isLoading, isAuthenticated, sessionExpired, route]);

  if (isLoading) {
    return (
      <div class="flex items-center justify-center h-screen bg-background-light dark:bg-background-dark">
        <span class="material-symbols-outlined text-5xl animate-spin text-primary">
          progress_activity
        </span>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}
