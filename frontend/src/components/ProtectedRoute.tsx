import { useLocation } from "preact-iso";
import { useEffect } from "preact/hooks";
import type { ComponentChildren } from "preact";
import { useAuth } from "../contexts/AuthContext";
import { LoadingSpinner } from "./LoadingSpinner";

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
    return (
      <LoadingSpinner className="h-dvh bg-background-light dark:bg-background-dark" />
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}
