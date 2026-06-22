import { startOfDay, addDays, addWeeks, addMonths, getDay } from "date-fns";

const WEEKDAYS: Record<string, number> = {
  sunday: 0, monday: 1, tuesday: 2, wednesday: 3,
  thursday: 4, friday: 5, saturday: 6,
};

export function parseNLDate(input: string): Date | null {
  const s = input.toLowerCase().trim();
  if (!s) return null;
  const today = startOfDay(new Date());

  if (s === "today") return today;
  if (s === "tomorrow") return addDays(today, 1);
  if (s === "yesterday") return addDays(today, -1);
  if (s === "next week") return addWeeks(today, 1);
  if (s === "next month") return addMonths(today, 1);
  if (s === "this weekend" || s === "weekend") {
    const day = getDay(today);
    return addDays(today, day <= 6 ? 6 - day : 1);
  }

  // "in X days/weeks/months"
  const inMatch = s.match(/^in (\d+) (day|days|week|weeks|month|months)$/);
  if (inMatch) {
    const n = parseInt(inMatch[1], 10);
    const unit = inMatch[2];
    if (unit.startsWith("day")) return addDays(today, n);
    if (unit.startsWith("week")) return addWeeks(today, n);
    if (unit.startsWith("month")) return addMonths(today, n);
  }

  // "next <weekday>"
  const nextDayMatch = s.match(/^(?:next )?(monday|tuesday|wednesday|thursday|friday|saturday|sunday)$/);
  if (nextDayMatch) {
    const target = WEEKDAYS[nextDayMatch[1]];
    const current = getDay(today);
    const daysUntil = target > current
      ? target - current
      : 7 - current + target;
    return addDays(today, daysUntil === 0 ? 7 : daysUntil);
  }

  return null;
}

export function formatNLHint(date: Date): string {
  const today = startOfDay(new Date());
  const diff = Math.round((date.getTime() - today.getTime()) / 86400000);
  if (diff === 0) return "Today";
  if (diff === 1) return "Tomorrow";
  if (diff === -1) return "Yesterday";
  if (diff > 0 && diff < 7) return date.toLocaleDateString("en-US", { weekday: "long" });
  return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}
