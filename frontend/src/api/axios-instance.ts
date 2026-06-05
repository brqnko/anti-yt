import Axios, { type InternalAxiosRequestConfig } from "axios";
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

type RetryableConfig = InternalAxiosRequestConfig & {
  _retry?: boolean;
  _gen?: number;
};

// Token-refresh coordination.
//
// `refreshGeneration` is bumped on every successful refresh. Each outgoing
// request is stamped (below) with the generation that was current when it was
// sent, so the response interceptor can tell whether a refresh already happened
// after a given request left the client.
//
// `refreshPromise` holds the single in-flight refresh so that every concurrent
// 401 awaits the same request instead of each firing its own. This matters
// because refresh tokens are single-use on the backend (rotation): a second,
// redundant refresh would be rejected and tear down the session.
let refreshGeneration = 0;
let refreshPromise: Promise<void> | null = null;

const performRefresh = (): Promise<void> => {
  if (!refreshPromise) {
    refreshPromise = axiosInstance
      .post("/api/v1/auth/refresh", undefined, { timeout: 15000 })
      .then(() => {
        refreshGeneration += 1;
      })
      .catch((refreshError) => {
        window.dispatchEvent(
          new CustomEvent("auth:logout", {
            detail: { reason: "session_expired" },
          }),
        );
        throw refreshError;
      })
      .finally(() => {
        refreshPromise = null;
      });
  }
  return refreshPromise;
};

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

  // Stamp the refresh generation current when this request leaves, so the
  // response interceptor can detect a refresh that completes while this request
  // is in flight and retry instead of starting a second (redundant) refresh.
  (config as RetryableConfig)._gen = refreshGeneration;

  return config;
});

// Response interceptor: automatic token refresh on 401
axiosInstance.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config as RetryableConfig | undefined;

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
      !originalRequest ||
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

    originalRequest._retry = true;

    // A refresh already completed after this request was sent (its stamped
    // generation is behind the current one), so the cookie is fresh now. Retry
    // directly instead of starting a redundant second refresh — which the
    // backend would reject because refresh tokens are single-use, tearing down
    // an otherwise valid session.
    if (
      typeof originalRequest._gen === "number" &&
      originalRequest._gen < refreshGeneration
    ) {
      return axiosInstance(originalRequest);
    }

    // Otherwise share the single in-flight refresh (or start one), then retry.
    try {
      await performRefresh();
      return axiosInstance(originalRequest);
    } catch (refreshError) {
      return Promise.reject(refreshError);
    }
  },
);
