import type { AxiosRequestConfig } from "axios";
import { axiosInstance } from "./axios-instance";

export const customInstance = <T>(config: AxiosRequestConfig): Promise<T> => {
  const controller = new AbortController();
  const promise = axiosInstance({
    ...config,
    signal: controller.signal,
  }).then(({ data }) => data);

  // @ts-expect-error -- orval expects cancel property on promise
  promise.cancel = () => {
    controller.abort();
  };

  return promise;
};
