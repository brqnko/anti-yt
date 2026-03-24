import { useState, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { getAuth } from "../../api/generated/auth";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import type { GetUsersMeSessions200ItemsItem } from "../../api/generated/antiYtApi.schemas";

function getDeviceIcon(deviceType: string): string {
  const lower = deviceType.toLowerCase();
  if (lower.includes("iphone") || lower.includes("android") || lower.includes("mobile")) {
    return "smartphone";
  }
  if (lower.includes("ipad") || lower.includes("tablet")) {
    return "tablet_mac";
  }
  return "desktop_windows";
}

function formatLastActive(dateStr: string, t: (key: string, opts?: Record<string, unknown>) => string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diffMs = now - then;
  const diffMin = Math.floor(diffMs / 60000);

  if (diffMin < 1) return t("security.justNow");
  if (diffMin < 60) return t("security.minutesAgo", { count: diffMin });
  const diffHours = Math.floor(diffMin / 60);
  if (diffHours < 24) return t("security.hoursAgo", { count: diffHours });
  const diffDays = Math.floor(diffHours / 24);
  if (diffDays < 30) return t("security.daysAgo", { count: diffDays });
  const diffMonths = Math.floor(diffDays / 30);
  return t("security.monthsAgo", { count: diffMonths });
}

const PAGE_SIZE = 20;

export function SecurityTab() {
  const { t } = useTranslation();
  const [sessions, setSessions] = useState<GetUsersMeSessions200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [revokingId, setRevokingId] = useState<string | null>(null);
  const [confirmRevokeId, setConfirmRevokeId] = useState<string | null>(null);
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [revokeError, setRevokeError] = useState<string | null>(null);

  const loadSessions = async (cursor?: string) => {
    try {
      setError(null);
      const { getUsersMeSessions } = getAuth();
      const data = await getUsersMeSessions({ limit: PAGE_SIZE, cursor });
      if (cursor) {
        setSessions((prev) => [...prev, ...data.items]);
      } else {
        setSessions(data.items);
        // TODO: 現在のセッションを正確に判定する
        if (data.items.length > 0) {
          setCurrentSessionId(data.items[0].id);
        }
      }
      setHasNext(data.has_next);
    } catch {
      setError(t("security.loadError"));
    } finally {
      setIsLoading(false);
      setIsLoadingMore(false);
    }
  };

  const loadMore = () => {
    if (sessions.length === 0 || !hasNext) return;
    setIsLoadingMore(true);
    const lastId = sessions[sessions.length - 1].id;
    loadSessions(lastId);
  };

  useEffect(() => {
    loadSessions();
  }, []);

  const handleRevoke = async (sessionId: string) => {
    setRevokingId(sessionId);
    setRevokeError(null);
    try {
      const { deleteUsersMeSessionsSessionId } = getAuth();
      await deleteUsersMeSessionsSessionId(sessionId);
      setSessions((prev) => prev.filter((s) => s.id !== sessionId));
      setConfirmRevokeId(null);
    } catch {
      setRevokeError(t("security.revokeError"));
    } finally {
      setRevokingId(null);
    }
  };

  const currentSession = sessions.find((s) => s.id === currentSessionId);
  const otherSessions = sessions.filter((s) => s.id !== currentSessionId);

  if (isLoading) {
    return <LoadingSpinner />;
  }

  return (
    <div class="flex flex-col gap-8">
      {/* Page Heading */}
      <div class="flex flex-col gap-2 mb-2">
        <h1 class="text-3xl lg:text-4xl font-black leading-tight tracking-[-0.033em]">
          {t("security.title")}
        </h1>
      </div>

      {/* Error banner */}
      {error && (
        <div class="flex items-center gap-3 px-4 py-3 rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 text-sm font-medium">
          <span class="material-symbols-outlined text-base">error</span>
          <span class="flex-1">{error}</span>
          <button
            onClick={() => { setError(null); setIsLoading(true); loadSessions(); }}
            class="text-red-600 dark:text-red-400 font-bold hover:underline cursor-pointer bg-transparent border-none text-sm"
          >
            {t("security.retry")}
          </button>
        </div>
      )}

      {/* Current Session */}
      {currentSession && (
        <div class="flex flex-col gap-4">
          <h3 class="text-lg font-bold leading-tight tracking-[-0.015em]">
            {t("security.currentSession")}
          </h3>
          <div class="overflow-hidden rounded-xl border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark shadow-sm">
            {/* Main Row */}
            <div class="flex flex-col sm:flex-row sm:items-center gap-4 p-6 border-b border-border-light dark:border-border-dark">
              <div class="flex items-center gap-4 flex-1">
                <div class="flex items-center justify-center rounded-lg bg-primary/20 shrink-0 size-14">
                  <span class="material-symbols-outlined text-2xl text-primary font-bold">
                    {getDeviceIcon(currentSession.device_type)}
                  </span>
                </div>
                <div class="flex flex-col justify-center">
                  <div class="flex items-center gap-2 flex-wrap">
                    <p class="text-lg font-bold leading-normal line-clamp-1">
                      {currentSession.browser_name}
                    </p>
                    <span class="bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-bold px-2 py-0.5 rounded-full border border-green-200 dark:border-green-800">
                      {t("security.activeNow")}
                    </span>
                  </div>
                  <p class="text-text-muted-light dark:text-text-muted-dark text-sm font-medium leading-normal mt-1">
                    {currentSession.device_type} · {currentSession.city_name}, {currentSession.country_code}
                  </p>
                </div>
              </div>
            </div>
            {/* Details */}
            <div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-4 bg-background-light/50 dark:bg-background-dark/30">
              <div class="flex flex-col gap-1">
                <p class="text-text-muted-light dark:text-text-muted-dark text-xs font-bold uppercase tracking-wider">
                  {t("security.loggedInAt")}
                </p>
                <p class="text-sm font-medium leading-relaxed">
                  {new Date(currentSession.created_at).toLocaleString()}
                </p>
              </div>
              <div class="flex flex-col gap-1">
                <p class="text-text-muted-light dark:text-text-muted-dark text-xs font-bold uppercase tracking-wider">
                  {t("security.lastActivity")}
                </p>
                <p class="text-sm font-medium leading-relaxed">
                  {formatLastActive(currentSession.last_logged_in_at, t)}
                </p>
              </div>
              <div class="flex flex-col gap-1">
                <p class="text-text-muted-light dark:text-text-muted-dark text-xs font-bold uppercase tracking-wider">
                  {t("security.ipAddress")}
                </p>
                <p class="text-sm font-medium leading-relaxed ">
                  {currentSession.ip_address}
                </p>
              </div>
              <div class="flex flex-col gap-1 md:col-span-2">
                <p class="text-text-muted-light dark:text-text-muted-dark text-xs font-bold uppercase tracking-wider">
                  {t("security.userAgent")}
                </p>
                <p class="text-sm font-medium leading-relaxed  break-all">
                  {currentSession.user_agent}
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Other Active Sessions */}
      {otherSessions.length > 0 && (
        <div class="flex flex-col gap-4">
          <h3 class="text-lg font-bold leading-tight tracking-[-0.015em]">
            {t("security.otherSessions")}
          </h3>
          <div class="flex flex-col gap-3">
            {otherSessions.map((session) => (
              <div
                key={session.id}
                class="flex flex-col sm:flex-row sm:items-center gap-4 bg-card-light dark:bg-card-dark px-5 py-4 rounded-xl border border-border-light dark:border-border-dark shadow-sm justify-between group hover:border-primary/50 transition-colors"
              >
                <div class="flex items-center gap-4">
                  <div class="flex items-center justify-center rounded-lg bg-background-light dark:bg-background-dark shrink-0 size-12">
                    <span class="material-symbols-outlined text-text-muted-light dark:text-text-muted-dark">
                      {getDeviceIcon(session.device_type)}
                    </span>
                  </div>
                  <div class="flex flex-col justify-center">
                    <p class="text-base font-bold leading-normal">
                      {session.browser_name}
                    </p>
                    <div class="flex items-center gap-2 text-text-muted-light dark:text-text-muted-dark text-sm flex-wrap">
                      <span>{session.device_type}</span>
                      <span class="size-1 rounded-full bg-text-muted-light dark:bg-text-muted-dark" />
                      <span>{session.city_name}, {session.country_code}</span>
                      <span class="size-1 rounded-full bg-text-muted-light dark:bg-text-muted-dark" />
                      <span class="">{session.ip_address}</span>
                      <span class="size-1 rounded-full bg-text-muted-light dark:bg-text-muted-dark" />
                      <span>{formatLastActive(session.last_logged_in_at, t)}</span>
                    </div>
                  </div>
                </div>
                <div class="flex items-center pl-16 sm:pl-0">
                  <button
                    onClick={() => setConfirmRevokeId(session.id)}
                    class="text-red-600 dark:text-red-400 text-sm font-bold hover:bg-red-50 dark:hover:bg-red-900/20 px-3 py-1.5 rounded-lg transition-colors cursor-pointer bg-transparent border-none"
                  >
                    {t("security.revoke")}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Load More */}
      {hasNext && (
        <div class="flex justify-center">
          <button
            onClick={loadMore}
            disabled={isLoadingMore}
            class="px-6 py-2.5 rounded-lg font-bold text-sm text-primary hover:bg-primary/10 transition-colors cursor-pointer bg-transparent border border-primary/30 disabled:opacity-50 flex items-center gap-2"
          >
            {isLoadingMore && (
              <span class="material-symbols-outlined text-[18px] animate-spin">
                progress_activity
              </span>
            )}
            {t("security.loadMore")}
          </button>
        </div>
      )}

      {/* Empty state */}
      {sessions.length === 0 && (
        <div class="text-center py-12 text-text-muted-light dark:text-text-muted-dark">
          <span class="material-symbols-outlined text-4xl mb-2">devices</span>
          <p class="font-medium">{t("security.noSessions")}</p>
        </div>
      )}

      {/* Revoke Confirmation Dialog */}
      {confirmRevokeId && (() => {
        const session = sessions.find((s) => s.id === confirmRevokeId);
        if (!session) return null;
        return (
          <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm" onClick={() => { if (revokingId === null) setConfirmRevokeId(null); }}>
            <div class="bg-card-light dark:bg-card-dark rounded-xl shadow-2xl border border-border-light dark:border-border-dark max-w-md w-full mx-4 p-6 flex flex-col gap-4" onClick={(e) => e.stopPropagation()}>
              <h3 class="text-lg font-bold">{t("security.revokeConfirmTitle")}</h3>
              <p class="text-sm text-text-muted-light dark:text-text-muted-dark leading-relaxed">
                {t("security.revokeConfirmDesc", { name: session.browser_name })}
              </p>
              {revokeError && (
                <div class="flex items-center gap-2 px-3 py-2 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 text-sm">
                  <span class="material-symbols-outlined text-base">error</span>
                  {revokeError}
                </div>
              )}
              <div class="flex justify-end gap-3 mt-2">
                <button
                  onClick={() => setConfirmRevokeId(null)}
                  disabled={revokingId !== null}
                  class="px-5 py-2.5 rounded-lg font-bold text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer bg-transparent border-none"
                >
                  {t("security.cancel")}
                </button>
                <button
                  onClick={() => handleRevoke(confirmRevokeId)}
                  disabled={revokingId !== null}
                  class="px-5 py-2.5 bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white font-bold rounded-lg transition-colors cursor-pointer border-none flex items-center gap-2"
                >
                  {revokingId === confirmRevokeId && (
                    <span class="material-symbols-outlined text-[18px] animate-spin">
                      progress_activity
                    </span>
                  )}
                  {t("security.revokeConfirm")}
                </button>
              </div>
            </div>
          </div>
        );
      })()}
    </div>
  );
}
