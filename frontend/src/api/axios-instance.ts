import Axios from "axios";
import { getCookie } from "../utils/cookie";

let cachedVisitorId: string | null = null;
let visitorIdPromise: Promise<string | null> | null = null;

export const getCachedVisitorId = (): string | null => cachedVisitorId;

const loadVisitorId = async (): Promise<string | null> => {
  try {
    const FingerprintJS = (await import("@fingerprintjs/fingerprintjs")).default;
    const fp = await FingerprintJS.load();
    const result = await fp.get();
    cachedVisitorId = result.visitorId;
    return cachedVisitorId;
  } catch {
    return null;
  }
};

const getVisitorId = (): Promise<string | null> => {
  if (typeof window === "undefined") return Promise.resolve(null);
  if (cachedVisitorId) return Promise.resolve(cachedVisitorId);
  if (!visitorIdPromise) {
    visitorIdPromise = loadVisitorId();
  }
  return visitorIdPromise;
};

if (typeof window !== "undefined") {
  const idle =
    (window as unknown as { requestIdleCallback?: (cb: () => void) => void })
      .requestIdleCallback ?? ((cb: () => void) => setTimeout(cb, 1));
  idle(() => {
    void getVisitorId();
  });
}

export const axiosInstance = Axios.create({
  withCredentials: true,
});

// Request interceptor: fingerprint + CSRF token
axiosInstance.interceptors.request.use(async (config) => {
  if (typeof window === "undefined") return config;

  const visitorId = await getVisitorId();
  if (visitorId) {
    config.headers["X-Device-Fingerprint"] = visitorId;
  }

  const csrfToken = getCookie("csrf_token");
  if (csrfToken) {
    config.headers["x-csrf-token"] = csrfToken;
  }

  config.headers["X-Timezone"] = Intl.DateTimeFormat().resolvedOptions().timeZone;

  return config;
});

// Response interceptor: automatic token refresh on 401
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: unknown | null) => {
  failedQueue.forEach((prom) => {
    if (error) prom.reject(error);
    else prom.resolve();
  });
  failedQueue = [];
};

axiosInstance.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // Handle 429 Too Many Requests (rate limit)
    if (error.response?.status === 429) {
      window.dispatchEvent(
        new CustomEvent("notification:show", {
          detail: { type: "error", messageKey: "apiErrors.too_many_requests" },
        }),
      );
      return Promise.reject(error);
    }

    // Handle 403 Forbidden (screen time restrictions)
    if (error.response?.status === 403) {
      const title: string = error.response?.data?.title ?? "";
      if (
        title === "screen_time_limit_exceeded" ||
        title === "outside_allowed_time_range"
      ) {
        const reason =
          title === "screen_time_limit_exceeded"
            ? "limit_exceeded"
            : "outside_time_range";
        window.dispatchEvent(
          new CustomEvent("screen-time:blocked", { detail: { reason } }),
        );
      }
      return Promise.reject(error);
    }

    if (
      error.response?.status !== 401 ||
      originalRequest._retry ||
      originalRequest.url === "/api/v1/auth/refresh"
    ) {
      return Promise.reject(error);
    }

    // No CSRF cookie means there's no session to refresh. Skip the refresh
    // request to avoid a guaranteed 400 surfacing as a console error.
    if (!getCookie("csrf_token")) {
      return Promise.reject(error);
    }

    // jti_blacklisted is raised during register/reactivation flows when a
    // one-time token has already been consumed. Refreshing the session token
    // would not help and would mask the original error, so surface it to the
    // caller directly.
    if (error.response?.data?.title === "jti_blacklisted") {
      return Promise.reject(error);
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      }).then(() => {
        originalRequest._retry = true;
        return axiosInstance(originalRequest);
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      await axiosInstance.post("/api/v1/auth/refresh", undefined, { timeout: 15000 });
      processQueue(null);
      return axiosInstance(originalRequest);
    } catch (refreshError) {
      processQueue(refreshError);
      window.dispatchEvent(
        new CustomEvent("auth:logout", {
          detail: { reason: "session_expired" },
        }),
      );
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);
