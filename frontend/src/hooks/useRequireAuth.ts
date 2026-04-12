import { useState, useCallback } from "preact/hooks";
import { useAuth } from "../contexts/AuthContext";

export function useRequireAuth() {
  const { isAuthenticated } = useAuth();
  const [showAuthPrompt, setShowAuthPrompt] = useState(false);

  const requireAuth = useCallback(
    (fn: () => void | Promise<void>): Promise<void> => {
      if (isAuthenticated) {
        return Promise.resolve(fn());
      }
      setShowAuthPrompt(true);
      return Promise.resolve();
    },
    [isAuthenticated],
  );

  const closeAuthPrompt = useCallback(() => setShowAuthPrompt(false), []);

  return { isAuthenticated, requireAuth, showAuthPrompt, closeAuthPrompt };
}
