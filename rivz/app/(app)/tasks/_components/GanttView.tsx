"use client";
import { differenceInDays, parseISO, startOfDay, addDays, isAfter, isBefore, format } from "date-fns";
import type { Task } from "@/lib/tasks-hooks";
import { cn } from "@/lib/utils";

const WINDOW_DAYS = 30;

const priorityColor: Record<string, string> = {
  high: "bg-rose-500",
  medium: "bg-amber-500",
  low: "bg-emerald-500",
};


type Props = {
  tasks: Task[];
  onTaskClick: (task: Task) => void;
};

export function GanttView({ tasks, onTaskClick }: Props) {
  const today = startOfDay(new Date());
  const windowEnd = addDays(today, WINDOW_DAYS);

  const visibleTasks = tasks
    .filter((t) => {
      if (!t.due_date) return false;
      const due = parseISO(t.due_date);
      // Include tasks due within window OR already overdue
      return !isAfter(due, windowEnd);
    })
    .sort((a, b) => {
      if (!a.due_date || !b.due_date) return 0;
      return parseISO(a.due_date).getTime() - parseISO(b.due_date).getTime();
    });

  if (visibleTasks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground gap-2">
        <span className="text-3xl">📅</span>
        <p className="text-sm">No tasks with due dates in the next {WINDOW_DAYS} days.</p>
      </div>
    );
  }

  // Day markers for the header
  const dayMarkers = [0, 7, 14, 21, 30].map((d) => ({
    label: d === 0 ? "Today" : `+${d}d`,
    pct: (d / WINDOW_DAYS) * 100,
  }));

  return (
    <div className="flex flex-col gap-1">
      {/* Header ruler */}
      <div className="ml-[200px] relative h-6 select-none mb-1">
        <div className="relative h-full">
          {dayMarkers.map(({ label, pct }) => (
            <span
              key={label}
              className="absolute text-[10px] text-muted-foreground -translate-x-1/2"
              style={{ left: `${pct}%` }}
            >
              {label}
            </span>
          ))}
        </div>
      </div>

      {/* Task rows */}
      <div className="flex flex-col gap-1.5">
        {visibleTasks.map((task) => {
          const due = parseISO(task.due_date!);
          const isOverdue = isBefore(due, today);

          // Bar starts at today (leftmost) and ends at due_date
          // For overdue tasks, bar goes from 0% to some small negative (cap at 0), show as past-due
          const daysUntilDue = differenceInDays(due, today);
          const barWidthPct = Math.max(2, Math.min(100, ((daysUntilDue + 1) / WINDOW_DAYS) * 100));
          const barLeftPct = 0;

          return (
            <div
              key={task.id}
              className="flex items-center gap-3 group cursor-pointer"
              onClick={() => onTaskClick(task)}
            >
              {/* Title column */}
              <div className="w-[200px] shrink-0 flex items-center gap-1.5 pr-3">
                <span
                  className={cn("size-2 rounded-full shrink-0", priorityColor[task.priority])}
                />
                <span className="text-xs font-medium truncate group-hover:text-foreground text-muted-foreground transition-colors">
                  {task.title}
                </span>
              </div>

              {/* Bar column */}
              <div className="flex-1 relative h-7 bg-muted/40 rounded-md overflow-hidden border border-border/50">
                {/* Grid lines */}
                {[25, 50, 75].map((pct) => (
                  <div
                    key={pct}
                    className="absolute top-0 bottom-0 w-px bg-border/40"
                    style={{ left: `${pct}%` }}
                  />
                ))}

                {/* Today marker */}
                <div className="absolute top-0 bottom-0 w-px bg-blue-500/60 z-10" style={{ left: "0%" }} />

                {/* Task bar */}
                <div
                  className={cn(
                    "absolute top-1 bottom-1 rounded flex items-center px-2 text-[10px] font-medium text-white min-w-[24px] transition-opacity group-hover:opacity-90",
                    isOverdue ? "bg-rose-500" : priorityColor[task.priority]
                  )}
                  style={{
                    left: `${barLeftPct}%`,
                    width: `${barWidthPct}%`,
                  }}
                >
                  <span className="truncate">
                    {isOverdue
                      ? `Overdue ${Math.abs(daysUntilDue)}d`
                      : daysUntilDue === 0
                      ? "Due today"
                      : `${daysUntilDue}d`}
                  </span>
                </div>
              </div>

              {/* Due date label */}
              <div
                className={cn(
                  "w-[68px] shrink-0 text-[10px] text-right",
                  isOverdue ? "text-rose-500 font-medium" : "text-muted-foreground"
                )}
              >
                {format(due, "MMM d")}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
