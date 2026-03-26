import { useLocation } from "preact-iso";
import { useEffect } from "preact/hooks";
import { lazy } from "preact-iso";
import type { ComponentChildren } from "preact";
import { useAuth } from "../contexts/AuthContext";
import { LoadingSpinner } from "./LoadingSpinner";

const ScreenTimeBlock = lazy(
  () => import("../pages/ScreenTimeBlock/index.tsx"),
);

interface ProtectedRouteProps {
  children: ComponentChildren;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const {
    isAuthenticated,
    isLoading,
    sessionExpired,
    screenTimeBlocked,
    screenTimeBlockReason,
  } = useAuth();
  const { route } = useLocation();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      route(sessionExpired ? "/?expired=1" : "/");
    }
  }, [isLoading, isAuthenticated, sessionExpired, route]);

  if (isLoading) {
    return (
      <LoadingSpinner className="h-dvh bg-background-light dark:bg-background-dark" />
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  if (screenTimeBlocked) {
    return (
      <ScreenTimeBlock reason={screenTimeBlockReason ?? "limit_exceeded"} />
    );
  }

  return <>{children}</>;
}
