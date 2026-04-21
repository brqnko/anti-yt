import { useLocation, lazy } from "preact-iso";
import type { ComponentChildren } from "preact";
import { useAuth } from "../contexts/AuthContext";

const ScreenTimeBlock = lazy(() => import("../pages/ScreenTimeBlock/index"));

interface ScreenTimeGateProps {
  children: ComponentChildren;
}

export function ScreenTimeGate({ children }: ScreenTimeGateProps) {
  const { isAuthenticated, screenTimeBlocked, screenTimeBlockReason } =
    useAuth();
  const { url } = useLocation();

  if (
    isAuthenticated &&
    screenTimeBlocked &&
    url !== "/screen-time-settings"
  ) {
    return (
      <ScreenTimeBlock reason={screenTimeBlockReason ?? "limit_exceeded"} />
    );
  }

  return <>{children}</>;
}
