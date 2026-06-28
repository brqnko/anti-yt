import { AxiosError } from "axios";

/**
 * Extracts the error code from an API error response's `title` field.
 * Returns null if the error is not an AxiosError or has no title.
 */
export function getApiErrorCode(err: unknown): string | null {
  if (err instanceof AxiosError && typeof err.response?.data?.title === "string") {
    return err.response.data.title;
  }
  return null;
}

/**
 * Resolves an i18n message key for an API error, for use with notifications.
 * Prefers the specific `apiErrors.<code>` key when it exists, otherwise
 * returns the given fallback key.
 */
export function apiErrorMessageKey(
  i18n: { exists: (key: string) => boolean },
  err: unknown,
  fallbackKey: string,
): string {
  const code = getApiErrorCode(err);
  if (code && i18n.exists(`apiErrors.${code}`)) return `apiErrors.${code}`;
  return fallbackKey;
}
