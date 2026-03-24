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
