import { useEffect } from "preact/hooks";

export function useTitle(title: string) {
  useEffect(() => {
    document.title = title ? `${title} | anti-yt` : "anti-yt";
  }, [title]);
}
