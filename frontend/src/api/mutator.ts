import type { AxiosRequestConfig } from "axios";
import { axiosInstance } from "./axios-instance";

type CancelablePromise<T> = Promise<T> & { cancel: () => void };

export const customInstance = <T>(config: AxiosRequestConfig): CancelablePromise<T> => {
  const controller = new AbortController();
  return Object.assign(
    axiosInstance({
      ...config,
      signal: controller.signal,
    }).then(({ data }) => data),
    { cancel: () => controller.abort() },
  );
};
