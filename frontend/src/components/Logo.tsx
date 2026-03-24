import { useAuth } from "../contexts/AuthContext";

export function Logo() {
  const { isAuthenticated } = useAuth();

  return (
    <a
      href={isAuthenticated ? "/dashboard" : "/"}
      class="no-underline text-charcoal dark:text-white"
    >
      <span class="text-xl font-bold tracking-tight">anti-yt</span>
    </a>
  );
}
