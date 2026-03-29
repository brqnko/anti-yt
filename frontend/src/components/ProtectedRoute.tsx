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
    screenTimeBlocked,
    screenTimeBlockReason,
  } = useAuth();
  const { route, url } = useLocation();

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

  if (screenTimeBlocked && url !== "/screen-time-settings") {
    return (
      <ScreenTimeBlock reason={screenTimeBlockReason ?? "limit_exceeded"} />
    );
  }

  return <>{children}</>;
}
