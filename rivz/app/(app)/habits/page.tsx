"use client";
import { useState } from "react";
import { format, subDays, eachDayOfInterval } from "date-fns";
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
import { Flame, Plus, Trash2, Target } from "lucide-react";
import { toast } from "sonner";

const COLORS = ["#22c55e", "#3b82f6", "#f59e0b", "#ef4444", "#a855f7", "#ec4899"];

function HabitRow({ habit }: { habit: Habit }) {
  const toggle = useToggleHabit();
  const del = useDeleteHabit();
  const today = new Date();
  const days = eachDayOfInterval({ start: subDays(today, 6), end: today });
  const from = format(subDays(today, 6), "yyyy-MM-dd");
  const to = format(today, "yyyy-MM-dd");
  const { data: logs } = useHabitLogs(habit.id, from, to);
  const doneSet = new Set((logs ?? []).map((l) => l.date));
  const color = habit.color ?? "#22c55e";

  return (
    <div className="flex items-center gap-4 px-5 py-4">
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
        {days.map((d) => {
          const key = format(d, "yyyy-MM-dd");
          const done = doneSet.has(key);
          return (
            <button
              key={key}
              title={format(d, "EEE d")}
              onClick={() => toggle.mutate({ id: habit.id, date: key })}
              className={cn(
                "size-7 rounded-md border text-[10px] flex items-center justify-center transition-all",
                done ? "border-transparent text-white" : "border-border hover:border-foreground/40 text-muted-foreground"
              )}
              style={done ? { background: color } : undefined}
            >
              {format(d, "EEEEE")}
            </button>
          );
        })}
      </div>

      <Button
        size="icon"
        variant="ghost"
        className="size-8 text-muted-foreground hover:text-destructive"
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
          Build streaks. Tap a day to mark it done.
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
        <div className="flex flex-col divide-y divide-border rounded-xl border border-border bg-card overflow-hidden">
          {habits!.map((h) => (
            <HabitRow key={h.id} habit={h} />
          ))}
        </div>
      )}
    </div>
  );
}
