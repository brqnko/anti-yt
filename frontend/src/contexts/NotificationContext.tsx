import { createContext, type ComponentChildren } from "preact";
import { useState, useEffect, useCallback, useContext, useMemo, useRef } from "preact/hooks";
import { NotificationHost } from "../components/NotificationHost";

type NotificationType = "error" | "info" | "success";

interface NotificationItem {
  id: string;
  type: NotificationType;
  messageKey: string;
  durationMs: number;
  isExiting?: boolean;
}

interface ShowNotificationInput {
  type?: NotificationType;
  messageKey: string;
  durationMs?: number;
}

interface NotificationState {
  notifications: NotificationItem[];
  show: (input: ShowNotificationInput) => void;
  dismiss: (id: string) => void;
}

const DEFAULT_DURATION_MS = 5000;

const NotificationContext = createContext<NotificationState>({
  notifications: [],
  show: () => {},
  dismiss: () => {},
});

export function NotificationProvider({ children }: { children: ComponentChildren }) {
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const timersRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

  const dismiss = useCallback((id: string) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, isExiting: true } : n)),
    );
    setTimeout(() => {
      const timer = timersRef.current.get(id);
      if (timer) {
        clearTimeout(timer);
        timersRef.current.delete(id);
      }
      setNotifications((prev) => prev.filter((n) => n.id !== id));
    }, 420);
  }, []);

  const scheduleDismiss = useCallback(
    (id: string, durationMs: number) => {
      const existing = timersRef.current.get(id);
      if (existing) clearTimeout(existing);
      const timer = setTimeout(() => dismiss(id), durationMs);
      timersRef.current.set(id, timer);
    },
    [dismiss],
  );

  const show = useCallback(
    (input: ShowNotificationInput) => {
      const type = input.type ?? "info";
      const durationMs = input.durationMs ?? DEFAULT_DURATION_MS;

      setNotifications((prev) => {
        const existing = prev.find((n) => n.messageKey === input.messageKey);
        if (existing) {
          scheduleDismiss(existing.id, durationMs);
          return prev;
        }
        const id =
          typeof crypto !== "undefined" && "randomUUID" in crypto
            ? crypto.randomUUID()
            : `n-${Date.now()}-${Math.random().toString(36).slice(2)}`;
        scheduleDismiss(id, durationMs);
        return [...prev, { id, type, messageKey: input.messageKey, durationMs }];
      });
    },
    [scheduleDismiss],
  );

  useEffect(() => {
    if (typeof window === "undefined") return;
    const handler = (e: Event) => {
      const detail = (e as CustomEvent).detail;
      if (!detail || typeof detail.messageKey !== "string") return;
      show({
        type: detail.type,
        messageKey: detail.messageKey,
        durationMs: detail.durationMs,
      });
    };
    window.addEventListener("notification:show", handler);
    return () => window.removeEventListener("notification:show", handler);
  }, [show]);

  useEffect(() => {
    const timers = timersRef.current;
    return () => {
      timers.forEach((t) => clearTimeout(t));
      timers.clear();
    };
  }, []);

  const value = useMemo(
    () => ({ notifications, show, dismiss }),
    [notifications, show, dismiss],
  );

  return (
    <NotificationContext.Provider value={value}>
      {children}
      <NotificationHost />
    </NotificationContext.Provider>
  );
}

export function useNotification(): NotificationState {
  return useContext(NotificationContext);
}
