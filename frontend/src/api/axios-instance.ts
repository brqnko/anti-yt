import Axios from "axios";
import FingerprintJS from "@fingerprintjs/fingerprintjs";
import { getCookie } from "../utils/cookie";

let cachedVisitorId: string | null = null;

export const getCachedVisitorId = (): string | null => cachedVisitorId;

const getVisitorId = async (): Promise<string | null> => {
  if (cachedVisitorId) return cachedVisitorId;
  try {
    const fp = await FingerprintJS.load();
    const result = await fp.get();
    cachedVisitorId = result.visitorId;
    return cachedVisitorId;
  } catch {
    return null;
  }
};

export const axiosInstance = Axios.create({
  withCredentials: true,
});

// Request interceptor: fingerprint + CSRF token
axiosInstance.interceptors.request.use(async (config) => {
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
