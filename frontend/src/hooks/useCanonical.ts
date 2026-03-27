import { useEffect } from "preact/hooks";
import { SITE_URL } from "../constants";

export function useCanonical(path: string) {
  useEffect(() => {
    let link = document.querySelector<HTMLLinkElement>('link[rel="canonical"]');
    if (!link) {
      link = document.createElement("link");
      link.rel = "canonical";
      document.head.appendChild(link);
    }
    link.href = `${SITE_URL}${path}`;
  }, [path]);
}
