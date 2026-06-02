import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { logger } from "@/lib/logger";

export type ThemeMode = "LIGHT" | "DARK" | "SYSTEM";

type ThemeContextValue = {
  isDarkMode: boolean;
  theme: ThemeMode;
  setTheme: (theme: ThemeMode) => void;
  toggleTheme: () => void;
};

const ThemeContext = createContext<ThemeContextValue | null>(null);
const THEME_STORAGE_KEY = "live-api.theme";

function getStoredTheme(): ThemeMode {
  try {
    const v = localStorage.getItem(THEME_STORAGE_KEY) as ThemeMode | null;
    if (v === "LIGHT" || v === "DARK" || v === "SYSTEM") return v;
  } catch (error) {
    logger.warn("Failed to read theme from localStorage", { error });
  }
  return "SYSTEM";
}

function getSystemPrefersDark(): boolean {
  return typeof window !== "undefined" && window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<ThemeMode>(getStoredTheme);
  const [systemPrefersDark, setSystemPrefersDark] = useState<boolean>(getSystemPrefersDark);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = (event: MediaQueryListEvent) => {
      setSystemPrefersDark(event.matches);
      logger.debug("System theme changed", { prefersDark: event.matches });
    };
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  const isDarkMode = theme === "SYSTEM" ? systemPrefersDark : theme === "DARK";

  useEffect(() => {
    document.documentElement.classList.toggle("dark", isDarkMode);
    logger.debug("Theme applied to document", { theme, isDarkMode });
  }, [isDarkMode, theme]);

  const setTheme = useCallback((next: ThemeMode) => {
    try {
      localStorage.setItem(THEME_STORAGE_KEY, next);
    } catch (error) {
      logger.error("Failed to save theme to localStorage", { error });
    }
    setThemeState(next);
    logger.info("Theme changed", { next });
  }, []);

  const toggleTheme = useCallback(() => {
    setTheme(theme === "DARK" ? "LIGHT" : "DARK");
  }, [theme, setTheme]);

  const value = useMemo<ThemeContextValue>(
    () => ({ isDarkMode, theme, setTheme, toggleTheme }),
    [isDarkMode, theme, setTheme, toggleTheme]
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useTheme must be used within a ThemeProvider");
  return ctx;
}
