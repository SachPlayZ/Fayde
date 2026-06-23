"use client";
import { useState } from "react";
import { format, subDays, eachDayOfInterval, isToday } from "date-fns";
import {
  useHabits,
  useCreateHabit,
  useDeleteHabit,
  useToggleHabit,
  useHabitLogs,
  type Habit,
} from "@/lib/habits-hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { Check, Flame, Plus, Trash2, Target } from "lucide-react";
import { toast } from "sonner";

const COLORS = ["#22c55e", "#3b82f6", "#f59e0b", "#ef4444", "#a855f7", "#ec4899"];

const today = new Date();
const DAYS = eachDayOfInterval({ start: subDays(today, 6), end: today });

function DayHeaders() {
  return (
    <div className="flex items-center gap-4 px-5 py-2 border-b border-border bg-muted/30">
      <div className="min-w-0 flex-1" />
      <div className="flex items-center gap-1.5">
        {DAYS.map((d) => {
          const tod = isToday(d);
          return (
            <div
              key={format(d, "yyyy-MM-dd")}
              className={cn(
                "size-7 flex flex-col items-center justify-center gap-px",
                tod ? "text-primary" : "text-muted-foreground"
              )}
            >
              <span className={cn("text-[9px] uppercase font-medium leading-none", tod && "font-bold")}>
                {format(d, "EEEEE")}
              </span>
              <span className={cn("text-[10px] leading-none", tod && "font-bold underline underline-offset-2")}>
                {format(d, "d")}
              </span>
            </div>
          );
        })}
      </div>
      <div className="size-8 shrink-0" />
    </div>
  );
}

function HabitRow({ habit, index }: { habit: Habit; index: number }) {
  const toggle = useToggleHabit();
  const del = useDeleteHabit();
  const from = format(subDays(today, 6), "yyyy-MM-dd");
  const to = format(today, "yyyy-MM-dd");
  const { data: logs } = useHabitLogs(habit.id, from, to);
  const doneSet = new Set((logs ?? []).map((l) => l.date));
  const color = habit.color ?? "#22c55e";

  return (
    <div
      className="flex items-center gap-4 px-5 py-4 hover:bg-muted/30 transition-colors animate-in fade-in-0 slide-in-from-bottom-1 duration-300"
      style={{ animationDelay: `${index * 40}ms`, animationFillMode: "both" }}
    >
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="size-2.5 rounded-full shrink-0" style={{ background: color }} />
          <span className="font-medium truncate">{habit.name}</span>
        </div>
        <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
          <span className="inline-flex items-center gap-1">
            <Flame className={cn("size-3.5", habit.current_streak > 0 && "text-orange-500")} />
            {habit.current_streak} day streak
          </span>
          <span>best {habit.longest_streak}</span>
        </div>
      </div>

      <div className="flex items-center gap-1.5">
        {DAYS.map((d) => {
          const key = format(d, "yyyy-MM-dd");
          const done = doneSet.has(key);
          const tod = isToday(d);
          return (
            <button
              key={key}
              title={done ? `${format(d, "EEE d")} — done (click to undo)` : `${format(d, "EEE d")} — click to mark done`}
              onClick={() => toggle.mutate({ id: habit.id, date: key })}
              className={cn(
                "size-7 rounded-md flex items-center justify-center transition-all duration-150",
                done
                  ? "text-white shadow-sm"
                  : tod
                  ? "border-2 border-dashed border-muted-foreground/50 hover:border-foreground/60 hover:bg-muted/60 text-muted-foreground"
                  : "border border-border hover:border-foreground/30 hover:bg-muted/40 text-muted-foreground/40"
              )}
              style={done ? { background: color } : undefined}
            >
              {done ? <Check className="size-3.5 stroke-[3]" /> : null}
            </button>
          );
        })}
      </div>

      <Button
        size="icon"
        variant="ghost"
        className="size-8 text-muted-foreground hover:text-destructive shrink-0"
        onClick={() =>
          del.mutate(habit.id, { onSuccess: () => toast.success("Habit deleted") })
        }
      >
        <Trash2 className="size-4" />
      </Button>
    </div>
  );
}

export default function HabitsPage() {
  const { data: habits, isLoading } = useHabits();
  const create = useCreateHabit();
  const [name, setName] = useState("");
  const [color, setColor] = useState(COLORS[0]);

  const handleCreate = () => {
    if (!name.trim()) return;
    create.mutate(
      { name: name.trim(), color },
      {
        onSuccess: () => {
          setName("");
          toast.success("Habit created");
        },
      }
    );
  };

  return (
    <div className="flex flex-col gap-6 max-w-3xl">
      <div>
        <h2 className="text-xl font-bold tracking-tight">Habits</h2>
        <p className="text-sm text-muted-foreground mt-0.5">
          Click any day to mark it done. Click again to undo.
        </p>
      </div>

      <div className="flex items-center gap-2 rounded-xl border border-border bg-card p-2">
        <div className="flex items-center gap-1 pl-1">
          {COLORS.map((c) => (
            <button
              key={c}
              onClick={() => setColor(c)}
              className={cn(
                "size-5 rounded-full transition-transform",
                color === c && "ring-2 ring-offset-2 ring-offset-card ring-foreground scale-110"
              )}
              style={{ background: c }}
            />
          ))}
        </div>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleCreate()}
          placeholder="New habit name…"
          className="flex-1 border-0 shadow-none focus-visible:ring-0"
        />
        <Button onClick={handleCreate} disabled={!name.trim() || create.isPending}>
          <Plus className="size-4 mr-1" /> Add
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-20 rounded-xl border border-border bg-card animate-pulse" />
          ))}
        </div>
      ) : (habits?.length ?? 0) === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 py-16 rounded-xl border border-border bg-card text-muted-foreground">
          <Target className="size-8" />
          <p className="text-sm">No habits yet. Add one above.</p>
        </div>
      ) : (
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <DayHeaders />
          <div className="divide-y divide-border">
            {habits!.map((h, i) => (
              <HabitRow key={h.id} habit={h} index={i} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
