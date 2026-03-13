import Axios from "axios";
import FingerprintJS from "@fingerprintjs/fingerprintjs";

let cachedVisitorId: string | null = null;

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

export const axiosInstance = Axios.create();

axiosInstance.interceptors.request.use(async (config) => {
  const visitorId = await getVisitorId();
  if (visitorId) {
    config.headers["X-Device-Fingerprint"] = visitorId;
  }
  return config;
});
