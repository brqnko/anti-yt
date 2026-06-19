import { useEffect } from "preact/hooks";
import { useLocation } from "preact-iso";

// Only these paths should be indexed by search engines. Every other route —
// /watch/<id>, /search, /channels/<id>, /playlists/<id> and the account-only
// pages — surfaces third-party YouTube content or private data, so it is marked
// noindex. This mirrors the X-Robots-Tag header set in nginx.conf and keeps the
// rendered <meta name="robots"> consistent for JS-rendering crawlers.
const INDEXABLE = new Set(["/", "/about", "/terms", "/privacy"]);

export function RobotsMeta() {
  const { path } = useLocation();

  useEffect(() => {
    if (typeof document === "undefined") return;
    const content = INDEXABLE.has(path) ? "index, follow" : "noindex, nofollow";
    let el = document.head.querySelector<HTMLMetaElement>('meta[name="robots"]');
    if (!el) {
      el = document.createElement("meta");
      el.setAttribute("name", "robots");
      document.head.appendChild(el);
    }
    el.content = content;
  }, [path]);

  return null;
}
