import { useState, useCallback } from "preact/hooks";
import { useAuth } from "../contexts/AuthContext";

export function useRequireAuth() {
  const { isAuthenticated, isLoading } = useAuth();
  const [showAuthPrompt, setShowAuthPrompt] = useState(false);

  const requireAuth = useCallback(
    (fn: () => void | Promise<void>): Promise<void> => {
      if (isLoading) {
        return Promise.resolve();
      }
      if (isAuthenticated) {
        return Promise.resolve(fn());
      }
      setShowAuthPrompt(true);
      return Promise.resolve();
    },
    [isAuthenticated, isLoading],
  );

  const closeAuthPrompt = useCallback(() => setShowAuthPrompt(false), []);

  return { isAuthenticated, isLoading, requireAuth, showAuthPrompt, closeAuthPrompt };
}
