import { useLocation } from "preact-iso";
import { useEffect } from "preact/hooks";
import type { ComponentChildren } from "preact";
import { useAuth } from "../contexts/AuthContext";

interface ProtectedRouteProps {
  children: ComponentChildren;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading } = useAuth();
  const { route } = useLocation();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      route("/");
    }
  }, [isLoading, isAuthenticated, route]);

  if (isLoading) {
    return null;
  }

  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}
