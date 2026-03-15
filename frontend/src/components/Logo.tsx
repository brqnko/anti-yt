import { useAuth } from "../contexts/AuthContext";

export function Logo() {
  const { isAuthenticated } = useAuth();

  return (
    <a
      href={isAuthenticated ? "/dashboard" : "/"}
      class="flex items-center gap-2 no-underline text-charcoal dark:text-white"
    >
      <span class="material-symbols-outlined text-3xl text-primary">
        timelapse
      </span>
      <span class="text-xl font-bold tracking-tight">anti-yt</span>
    </a>
  );
}
