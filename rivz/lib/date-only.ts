import { formatDistanceToNow, isPast } from "date-fns";

// Combine a yyyy-MM-dd date with an optional "HH:mm" local time into a
// correct UTC instant. No time chosen -> defaults to end of day (23:59:59
// local), matching "not overdue until the day is over".
export function toDueInstant(dateStr: string, time?: string | null): string {
  const timePart = time && /^\d{2}:\d{2}$/.test(time) ? `${time}:00` : "23:59:59";
  return new Date(`${dateStr}T${timePart}`).toISOString();
}

export function isOverdue(iso: string | null | undefined): boolean {
  if (!iso) return false;
  return isPast(new Date(iso));
}

export function formatDueRelative(iso: string): string {
  const due = new Date(iso);
  return isPast(due)
    ? `Overdue by ${formatDistanceToNow(due)}`
    : `Due in ${formatDistanceToNow(due)}`;
}
