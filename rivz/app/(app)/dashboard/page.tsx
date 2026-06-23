"use client";
import { useState } from "react";
import Link from "next/link";
import { format, parseISO } from "date-fns";
import { useDashboard, type TaskBrief } from "@/lib/dashboard-hooks";
import { usePlanDay } from "@/lib/ai-hooks";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import {
  Sun,
  AlertCircle,
  CalendarClock,
  CheckCircle2,
  Timer,
  Flame,
  Clock,
  Plus,
  ArrowRight,
  Sparkles,
} from "lucide-react";

const PRIORITY_COLOR: Record<string, string> = {
  high: "text-rose-500",
  medium: "text-amber-500",
  low: "text-blue-500",
};

function Stat({
  icon,
  label,
  value,
  tint,
}: {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  tint: string;
}) {
  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <div className={cn("flex items-center gap-1.5 text-xs font-medium", tint)}>
        {icon}
        {label}
      </div>
      <p className="mt-2 text-2xl font-bold tabular-nums">{value}</p>
    </div>
  );
}

function TaskList({ tasks, empty }: { tasks: TaskBrief[]; empty: string }) {
  if (tasks.length === 0) {
    return <p className="px-4 py-6 text-sm text-muted-foreground text-center">{empty}</p>;
  }
  return (
    <div className="divide-y divide-border">
      {tasks.slice(0, 6).map((t) => (
        <Link
          key={t.id}
          href={`/tasks/${t.id}`}
          className="flex items-center gap-2 px-4 py-2.5 text-sm hover:bg-muted/50 transition-colors"
        >
          <span className={cn("size-1.5 rounded-full bg-current shrink-0", PRIORITY_COLOR[t.priority])} />
          <span className="truncate flex-1">{t.title}</span>
          {t.due_date && (
            <span className="text-xs text-muted-foreground shrink-0">
              {format(parseISO(t.due_date), "MMM d")}
            </span>
          )}
        </Link>
      ))}
    </div>
  );
}

function Panel({
  title,
  icon,
  href,
  children,
}: {
  title: string;
  icon: React.ReactNode;
  href: string;
  children: React.ReactNode;
}) {
  return (
    <div className="rounded-xl border border-border bg-card overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2.5 border-b border-border">
        <span className="flex items-center gap-1.5 text-sm font-semibold">
          {icon}
          {title}
        </span>
        <Link href={href} className="text-muted-foreground hover:text-foreground">
          <ArrowRight className="size-4" />
        </Link>
      </div>
      {children}
    </div>
  );
}

function PlanMyDay() {
  const planDay = usePlanDay();
  const [open, setOpen] = useState(false);

  const run = () => {
    setOpen(true);
    planDay.mutate({});
  };

  return (
    <>
      <Button variant="outline" size="sm" onClick={run} className="gap-1.5">
        <Sparkles className="size-4 text-violet-500" />
        Plan my day
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Sparkles className="size-4 text-violet-500" /> Your plan for today
            </DialogTitle>
          </DialogHeader>
          {planDay.isPending ? (
            <div className="space-y-2 py-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-6 rounded bg-muted animate-pulse" />
              ))}
            </div>
          ) : planDay.isError ? (
            <p className="text-sm text-muted-foreground py-4">
              Couldn&apos;t generate a plan. Make sure AI is configured and you have open tasks.
            </p>
          ) : (
            <pre className="text-sm whitespace-pre-wrap font-sans max-h-[60vh] overflow-y-auto">
              {planDay.data?.plan}
            </pre>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}

export default function DashboardPage() {
  const { data, isLoading } = useDashboard();

  if (isLoading || !data) {
    return (
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="h-24 rounded-xl border border-border bg-card animate-pulse" />
        ))}
      </div>
    );
  }

  const hours = Math.floor(data.time_this_week_minutes / 60);
  const mins = data.time_this_week_minutes % 60;

  return (
    <div className="flex flex-col gap-5">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight">Good day 👋</h2>
          <p className="text-sm text-muted-foreground mt-0.5">Here&apos;s your snapshot.</p>
        </div>
        <PlanMyDay />
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <Stat icon={<Sun className="size-3.5" />} label="Due today" value={data.due_today.length} tint="text-amber-500" />
        <Stat icon={<AlertCircle className="size-3.5" />} label="Overdue" value={data.overdue.length} tint="text-rose-500" />
        <Stat icon={<CheckCircle2 className="size-3.5" />} label="Done this week" value={data.completed_this_week} tint="text-emerald-500" />
        <Stat icon={<Clock className="size-3.5" />} label="Tracked this week" value={`${hours}h ${mins}m`} tint="text-blue-500" />
      </div>

      <div className="grid md:grid-cols-2 gap-3">
        <Panel title="Due today" icon={<Sun className="size-4 text-amber-500" />} href="/tasks?list=today">
          <TaskList tasks={data.due_today} empty="Nothing due today 🎉" />
        </Panel>
        <Panel title="Overdue" icon={<AlertCircle className="size-4 text-rose-500" />} href="/tasks?list=overdue">
          <TaskList tasks={data.overdue} empty="No overdue tasks" />
        </Panel>
        <Panel title="Upcoming" icon={<CalendarClock className="size-4 text-violet-500" />} href="/tasks?list=upcoming">
          <TaskList tasks={data.upcoming} empty="Nothing scheduled" />
        </Panel>
        <Panel title="Habits" icon={<Flame className="size-4 text-orange-500" />} href="/habits">
          {data.habits.length === 0 ? (
            <Link href="/habits" className="flex items-center gap-2 px-4 py-6 text-sm text-muted-foreground hover:text-foreground justify-center">
              <Plus className="size-4" /> Start a habit
            </Link>
          ) : (
            <div className="divide-y divide-border">
              {data.habits.slice(0, 6).map((h) => (
                <div key={h.id} className="flex items-center gap-2 px-4 py-2.5 text-sm">
                  <span className="size-2 rounded-full shrink-0" style={{ background: h.color ?? "#22c55e" }} />
                  <span className="truncate flex-1">{h.name}</span>
                  <span className={cn("flex items-center gap-1 text-xs", h.done_today ? "text-emerald-500" : "text-muted-foreground")}>
                    <Flame className="size-3" />
                    {h.current_streak}
                  </span>
                </div>
              ))}
            </div>
          )}
        </Panel>
      </div>

      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <Timer className="size-3.5" />
        {data.pomodoros_today} pomodoros today · {data.created_this_week} tasks created this week
      </div>
    </div>
  );
}
