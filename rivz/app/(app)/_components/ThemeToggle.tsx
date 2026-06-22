"use client";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { Moon, Sun } from "lucide-react";
import { useUpdatePreferences } from "@/lib/user-hooks";

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const updatePrefs = useUpdatePreferences();

  const toggle = () => {
    const next = theme === "dark" ? "light" : "dark";
    setTheme(next);
    updatePrefs.mutate({ theme: next });
  };

  return (
    <Button variant="ghost" size="icon" onClick={toggle}>
      <Sun className="rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      <Moon className="absolute rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
      <span className="sr-only">Toggle theme</span>
    </Button>
  );
}
