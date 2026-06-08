import { Component, type ComponentChildren } from "preact";

const RELOAD_GUARD_KEY = "chunk-reload-at";

// A failed dynamic import (a stale chunk after a deploy) reports a different
// message per engine — Safari/WebKit uses "Importing a module script failed".
function isChunkLoadError(error: unknown): boolean {
  const msg = error instanceof Error ? error.message : String(error ?? "");
  return /dynamically imported module|importing a module script failed|failed to fetch dynamically|loading chunk|chunkloaderror|unable to preload/i.test(
    msg,
  );
}

// Reload once to recover from a stale chunk. Returns false (and does nothing)
// if a reload already happened recently, so a persistent error never loops.
export function reloadForStaleChunk(): boolean {
  if (typeof window === "undefined") return false;
  try {
    const last = Number(sessionStorage.getItem(RELOAD_GUARD_KEY) || "0");
    if (Date.now() - last < 10_000) return false;
    sessionStorage.setItem(RELOAD_GUARD_KEY, String(Date.now()));
  } catch {
    // sessionStorage blocked (private mode etc.) — fall through and reload once.
  }
  window.location.reload();
  return true;
}

interface ErrorBoundaryState {
  hasError: boolean;
}

export class ErrorBoundary extends Component<
  { children: ComponentChildren },
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(error: unknown) {
    // The most common cause is a stale lazy chunk after a deploy: the old
    // hashed file is gone, so import() rejects. Reload to fetch the fresh shell.
    if (isChunkLoadError(error)) reloadForStaleChunk();
  }

  render() {
    if (this.state.hasError) {
      return (
        <div class="min-h-dvh flex flex-col items-center justify-center gap-4 p-6 text-center bg-background-light dark:bg-background-dark text-charcoal dark:text-white">
          <p class="text-lg font-medium">ページの読み込みに失敗しました</p>
          <button
            class="px-4 py-2 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary/90 cursor-pointer border-none"
            onClick={() => window.location.reload()}
          >
            再読み込み
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
