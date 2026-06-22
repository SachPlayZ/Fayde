"use client";
import { useMemo } from "react";
import { useRouter } from "next/navigation";
import { format, isBefore, isSameDay, parseISO, startOfDay } from "date-fns";
import { useTasks, useUpdateTask, type Task } from "@/lib/tasks-hooks";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { CheckCircle2, CalendarClock, SkipForward, ArrowRight } from "lucide-react";
import { toast } from "sonner";

const statusConfig = {
  todo: { label: "Todo", className: "bg-muted text-muted-foreground" },
  in_progress: { label: "In Progress", className: "bg-blue-500/10 text-blue-600 dark:text-blue-400" },
  done: { label: "Done", className: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400" },
  failed: { label: "Failed", className: "bg-rose-500/10 text-rose-600 dark:text-rose-400" },
};

const priorityDot: Record<string, string> = {
  low: "bg-emerald-500",
  medium: "bg-amber-500",
  high: "bg-rose-500",
};

function ReviewTaskCard({
  task,
  mode,
}: {
  task: Task;
  mode: "overdue" | "today";
}) {
  const router = useRouter();
  const updateTask = useUpdateTask();
  const today = startOfDay(new Date());

  const handleDoToday = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await updateTask.mutateAsync({
        id: task.id,
        due_date: `${format(today, "yyyy-MM-dd")}T00:00:00Z`,
      });
      toast.success("Moved to today");
    } catch {
      toast.error("Failed");
    }
  };

  const handleSkip = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await updateTask.mutateAsync({ id: task.id, due_date: null });
      toast.success("Due date cleared");
    } catch {
      toast.error("Failed");
    }
  };

  const handleDone = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await updateTask.mutateAsync({
        id: task.id,
        status: task.status === "done" ? "todo" : "done",
      });
    } catch {
      toast.error("Failed");
    }
  };

  const isDone = task.status === "done";

  return (
    <div
      className={cn(
        "group flex items-center gap-3 rounded-xl border border-border bg-card px-4 py-3 hover:shadow-sm transition-all cursor-pointer",
        isDone && "opacity-60"
      )}
      onClick={() => router.push(`/tasks/${task.id}`)}
    >
      <button
        type="button"
        onClick={handleDone}
        className="shrink-0 text-muted-foreground hover:text-emerald-500 transition-colors"
        aria-label={isDone ? "Mark undone" : "Mark done"}
      >
        <CheckCircle2 className={cn("size-5", isDone && "text-emerald-500")} />
      </button>

      <div className="flex-1 min-w-0">
        <p className={cn("text-sm font-medium truncate", isDone && "line-through text-muted-foreground")}>
          {task.title}
        </p>
        <div className="flex items-center gap-2 mt-0.5">
          <span className={cn("size-1.5 rounded-full shrink-0", priorityDot[task.priority])} />
          <span className={cn("text-xs px-1.5 py-0.5 rounded-full font-medium", statusConfig[task.status].className)}>
            {statusConfig[task.status].label}
          </span>
          {task.due_date && (
            <span className="text-xs text-muted-foreground">
              Due {format(parseISO(task.due_date), "MMM d")}
            </span>
          )}
        </div>
      </div>

      <div className="flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity" onClick={(e) => e.stopPropagation()}>
        {mode === "overdue" && (
          <>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="h-7 text-xs gap-1"
              onClick={handleDoToday}
            >
              <CalendarClock className="size-3" />
              Do today
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="h-7 text-xs gap-1 text-muted-foreground"
              onClick={handleSkip}
            >
              <SkipForward className="size-3" />
              Skip
            </Button>
          </>
        )}
        <ArrowRight className="size-4 text-muted-foreground" />
      </div>
    </div>
  );
}

export function DailyReviewClient() {
  const { data } = useTasks({ limit: 200 });
  const tasks = data?.data ?? [];

  const today = startOfDay(new Date());

  const { overdue, todayTasks } = useMemo(() => {
    const overdue = tasks.filter(
      (t) =>
        t.due_date &&
        isBefore(parseISO(t.due_date), today) &&
        t.status !== "done"
    );
    const todayTasks = tasks.filter(
      (t) => t.due_date && isSameDay(parseISO(t.due_date), today)
    );
    return { overdue, todayTasks };
  }, [tasks, today]);

  const doneToday = todayTasks.filter((t) => t.status === "done").length;

  return (
    <div className="max-w-2xl mx-auto flex flex-col gap-8">
      {/* Header */}
      <div>
        <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground mb-1">
          Daily Review
        </p>
        <h1 className="text-2xl font-bold">
          {format(new Date(), "EEEE, MMMM d")}
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          {todayTasks.length === 0
            ? "Nothing scheduled for today."
            : doneToday === todayTasks.length
            ? "All done for today!"
            : `${doneToday} of ${todayTasks.length} tasks done today`}
          {overdue.length > 0 && ` · ${overdue.length} overdue`}
        </p>
      </div>

      {/* Overdue */}
      {overdue.length > 0 && (
        <section>
          <h2 className="text-sm font-semibold text-rose-600 dark:text-rose-400 mb-3 flex items-center gap-2">
            <span className="size-2 rounded-full bg-rose-500 inline-block" />
            Overdue ({overdue.length})
          </h2>
          <div className="flex flex-col gap-2">
            {overdue.map((task) => (
              <ReviewTaskCard key={task.id} task={task} mode="overdue" />
            ))}
          </div>
        </section>
      )}

      {/* Today */}
      <section>
        <h2 className="text-sm font-semibold text-foreground mb-3 flex items-center gap-2">
          <span className="size-2 rounded-full bg-amber-500 inline-block" />
          Today ({todayTasks.length})
        </h2>
        {todayTasks.length === 0 ? (
          <p className="text-sm text-muted-foreground py-6 text-center border border-dashed border-border rounded-xl">
            Nothing scheduled for today. Enjoy your day!
          </p>
        ) : (
          <div className="flex flex-col gap-2">
            {todayTasks.map((task) => (
              <ReviewTaskCard key={task.id} task={task} mode="today" />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
