import { Moon, Sun } from "lucide-react";
import { useTheme } from "@/lib/use-theme";
import { cn } from "@/lib/cn";

export function ThemeToggle() {
  const { isDarkMode, toggleTheme } = useTheme();

  return (
    <button
      type="button"
      aria-label="Toggle theme"
      onClick={toggleTheme}
      className={cn(
        "relative inline-flex size-9 cursor-pointer items-center justify-center rounded-full",
        "border border-border bg-bg-elevated text-fg-muted",
        "transition-all duration-300 ease-[var(--ease-out-soft)]",
        "hover:text-fg hover:bg-bg-muted"
      )}
    >
      <Sun
        className={cn(
          "absolute size-4 transition-all duration-500 ease-[var(--ease-out-soft)]",
          isDarkMode ? "rotate-90 scale-0 opacity-0" : "rotate-0 scale-100 opacity-100"
        )}
      />
      <Moon
        className={cn(
          "absolute size-4 transition-all duration-500 ease-[var(--ease-out-soft)]",
          isDarkMode ? "rotate-0 scale-100 opacity-100" : "-rotate-90 scale-0 opacity-0"
        )}
      />
    </button>
  );
}
