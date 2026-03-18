interface LoadingSpinnerProps {
  className?: string;
  size?: "sm" | "lg";
}

export function LoadingSpinner({
  className = "py-20",
  size = "lg",
}: LoadingSpinnerProps) {
  return (
    <div class={`flex items-center justify-center ${className}`}>
      <span
        class={`material-symbols-outlined animate-spin text-primary ${size === "lg" ? "text-5xl" : "text-3xl"}`}
      >
        progress_activity
      </span>
    </div>
  );
}
