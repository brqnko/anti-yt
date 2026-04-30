import { useState, useEffect, useRef } from "preact/hooks";
import { createPortal } from "preact/compat";
import { useTranslation } from "react-i18next";
import { getAuth } from "../../api/generated/auth";
import { LoadingSpinner } from "../../components/LoadingSpinner";
import type { GetUsersMeSessions200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";

const PAGE_SIZE = 20;

function SessionMenu({ session, onRevoke }: { session: GetUsersMeSessions200ItemsItem; onRevoke: () => void }) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [menuPos, setMenuPos] = useState({ top: 0, right: 0 });
  const btnRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (btnRef.current && !btnRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  const handleOpen = () => {
    if (btnRef.current) {
      const rect = btnRef.current.getBoundingClientRect();
      setMenuPos({
        top: rect.bottom + window.scrollY + 4,
        right: window.innerWidth - rect.right,
      });
    }
    setOpen((v) => !v);
  };

  return (
    <>
      <button
        ref={btnRef}
        onClick={handleOpen}
        class="text-text-muted-light dark:text-text-muted-dark hover:text-foreground-light dark:hover:text-foreground-dark transition-colors cursor-pointer bg-transparent border-none p-1 rounded-lg hover:bg-black/5 dark:hover:bg-white/5"
        aria-label={t("security.revoke")}
      >
        <Icon name="more_vert" class="text-xl" />
      </button>
      {open && createPortal(
        <div
          style={{ position: "absolute", top: menuPos.top, right: menuPos.right }}
          class="z-[200] bg-card-light dark:bg-card-dark rounded-lg shadow-lg border border-border-light dark:border-border-dark p-1 min-w-[140px]"
        >
          <button
            onClick={() => { setOpen(false); onRevoke(); }}
            class="w-full text-left px-3 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 font-medium cursor-pointer bg-transparent border-none rounded-md"
          >
            {t("security.logout")}
          </button>
        </div>,
        document.body,
      )}
    </>
  );
}

export function SecurityTab() {
  const { t } = useTranslation();
  const [sessions, setSessions] = useState<GetUsersMeSessions200ItemsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasNext, setHasNext] = useState(false);
  const [revokingId, setRevokingId] = useState<string | null>(null);
  const [confirmRevokeId, setConfirmRevokeId] = useState<string | null>(null);
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

  if (isLoading) {
    return <LoadingSpinner />;
  }

  return (
    <div class="flex flex-col gap-8">
      {/* Error banner */}
      {error && (
        <div class="flex items-center gap-3 px-4 py-3 rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 text-sm font-medium">
          <Icon name="error" class="text-base" />
          <span class="flex-1">{error}</span>
          <button
            onClick={() => { setError(null); setIsLoading(true); loadSessions(); }}
            class="text-red-600 dark:text-red-400 font-bold hover:underline cursor-pointer bg-transparent border-none text-sm"
          >
            {t("security.retry")}
          </button>
        </div>
      )}

      {/* Active Sessions Table */}
      {sessions.length > 0 && (
        <div class="flex flex-col gap-4">
          <h3 class="text-lg font-bold leading-tight tracking-[-0.015em]">
            {t("security.activeSessions")}
          </h3>
          <div class="overflow-x-auto rounded-xl border border-border-light dark:border-border-dark">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-border-light dark:border-border-dark bg-background-light/50 dark:bg-background-dark/30">
                  <th class="text-left px-4 py-3 font-semibold text-text-muted-light dark:text-text-muted-dark whitespace-nowrap">
                    {t("security.device")}
                  </th>
                  <th class="text-left px-4 py-3 font-semibold text-text-muted-light dark:text-text-muted-dark whitespace-nowrap">
                    {t("security.location")}
                  </th>
                  <th class="text-left px-4 py-3 font-semibold text-text-muted-light dark:text-text-muted-dark whitespace-nowrap">
                    {t("security.createdAt")}
                  </th>
                  <th class="text-left px-4 py-3 font-semibold text-text-muted-light dark:text-text-muted-dark whitespace-nowrap">
                    {t("security.updatedAt")}
                  </th>
                  <th class="px-4 py-3 w-10" />
                </tr>
              </thead>
              <tbody>
                {sessions.map((session) => (
                  <tr
                    key={session.id}
                    class="border-b border-border-light dark:border-border-dark last:border-0 hover:bg-black/[0.03] dark:hover:bg-white/[0.03] transition-colors"
                  >
                    <td class="px-4 py-3 font-bold leading-normal">{session.browser_name}</td>
                    <td class="px-4 py-3 text-text-muted-light dark:text-text-muted-dark max-w-[200px] truncate">
                      {session.city_name}, {session.country_code}
                    </td>
                    <td class="px-4 py-3 text-text-muted-light dark:text-text-muted-dark whitespace-nowrap">
                      {new Date(session.created_at).toLocaleString()}
                    </td>
                    <td class="px-4 py-3 text-text-muted-light dark:text-text-muted-dark whitespace-nowrap">
                      {new Date(session.last_logged_in_at).toLocaleString()}
                    </td>
                    <td class="px-4 py-3 text-right">
                      <SessionMenu
                        session={session}
                        onRevoke={() => setConfirmRevokeId(session.id)}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
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
              <Icon name="progress_activity" class="text-[18px] animate-spin" />
            )}
            {t("security.loadMore")}
          </button>
        </div>
      )}

      {/* Empty state */}
      {sessions.length === 0 && (
        <div class="text-center py-12 text-text-muted-light dark:text-text-muted-dark">
          <Icon name="devices" class="text-4xl mb-2" />
          <p class="font-medium">{t("security.noSessions")}</p>
        </div>
      )}

      {/* Revoke Confirmation Dialog */}
      {confirmRevokeId && (() => {
        const session = sessions.find((s) => s.id === confirmRevokeId);
        if (!session) return null;
        return (
          <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60" onClick={() => { if (revokingId === null) setConfirmRevokeId(null); }}>
            <div class="bg-card-light dark:bg-card-dark rounded-xl ring-1 ring-black/10 dark:ring-white/10 border border-border-light dark:border-border-dark max-w-md w-full mx-4 p-6 flex flex-col gap-4" onClick={(e) => e.stopPropagation()}>
              <h3 class="text-lg font-bold">{t("security.revokeConfirmTitle")}</h3>
              <p class="text-sm text-text-muted-light dark:text-text-muted-dark leading-relaxed">
                {t("security.revokeConfirmDesc", { name: session.browser_name })}
              </p>
              {revokeError && (
                <div class="flex items-center gap-2 px-3 py-2 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 text-sm">
                  <Icon name="error" class="text-base" />
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
                    <Icon name="progress_activity" class="text-[18px] animate-spin" />
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
