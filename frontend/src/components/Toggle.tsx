export function Toggle({
  checked,
  disabled,
  onClick,
}: {
  checked: boolean;
  disabled?: boolean;
  onClick: () => void;
}) {
  return (
    <button
      class="relative inline-flex items-center cursor-pointer bg-transparent border-none p-0 flex-shrink-0"
      onClick={onClick}
      disabled={disabled}
    >
      <div
        class={`w-14 h-7 rounded-full transition-colors duration-200 ${
          checked ? "bg-primary" : "bg-gray-200 dark:bg-gray-700"
        } ${disabled ? "opacity-50" : ""}`}
      >
        <div
          class={`absolute top-0.5 left-[4px] bg-white border border-gray-300 rounded-full h-6 w-6 transition-transform duration-200 ${
            checked ? "translate-x-full" : ""
          }`}
        />
      </div>
    </button>
  );
}
